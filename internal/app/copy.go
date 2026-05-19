package app

import (
	"context"
	"log/slog"
	"sync"

	"github.com/kstieger/mongocp/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CopyCollections copies all collections for the given databases in parallel.
func CopyCollections(ctx context.Context, srcClient, dstClient *mongo.Client, dbs []domain.Database, workerCount int, dryRun bool, progressEnabled bool, logger *slog.Logger) error {
	tasks := make(chan domain.CopyTask)
	var wg sync.WaitGroup
	progress := newWorkerProgress(progressEnabled, workerCount)

	// Start workers
	for i := range workerCount {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range tasks {
				processCopyTask(ctx, srcClient, dstClient, task, workerID, dryRun, logger, progress)
			}
		}(i)
	}

	// Enqueue tasks
	for _, db := range dbs {
		colls, err := srcClient.Database(db.Name).ListCollectionNames(ctx, bson.D{})
		if err != nil {
			logger.Error("Failed to list collections", "db", db.Name, "err", err)
			continue
		}
		for _, coll := range colls {
			count, _ := srcClient.Database(db.Name).Collection(coll).CountDocuments(ctx, bson.D{})
			tasks <- domain.CopyTask{SrcDB: db.Name, DstDB: db.Name, Coll: coll, DocCount: count}
		}
	}
	close(tasks)
	wg.Wait()
	progress.Wait()
	return nil
}

func processCopyTask(
	ctx context.Context,
	srcClient, dstClient *mongo.Client,
	task domain.CopyTask,
	workerID int,
	dryRun bool,
	logger *slog.Logger,
	progress workerProgress,
) {
	progress.StartTask(workerID, task.SrcDB, task.Coll, task.DocCount)
	defer progress.CompleteTask(workerID)

	if dryRun {
		progress.Advance(workerID, task.DocCount)
		return
	}

	logger.Info("Copying collection", "db", task.SrcDB, "coll", task.Coll)
	srcColl := srcClient.Database(task.SrcDB).Collection(task.Coll)
	dstColl := dstClient.Database(task.DstDB).Collection(task.Coll)
	_ = dstColl.Drop(ctx)

	batchSize := 100
	cursor, err := srcColl.Find(ctx, bson.D{})
	if err != nil {
		logger.Error("Failed to read source collection", "db", task.SrcDB, "coll", task.Coll, "err", err)
		return
	}

	totalCopied := 0
	batch := make([]any, 0, batchSize)
	for cursor.Next(ctx) {
		var doc map[string]any
		if err := cursor.Decode(&doc); err != nil {
			logger.Error("Failed to decode document", "db", task.SrcDB, "coll", task.Coll, "err", err)
			continue
		}
		batch = append(batch, doc)
		if len(batch) >= batchSize {
			_, err := dstColl.InsertMany(ctx, batch)
			if err != nil {
				logger.Error("Failed to insert batch", "db", task.DstDB, "coll", task.Coll, "err", err)
				break
			}
			totalCopied += len(batch)
			progress.Advance(workerID, int64(len(batch)))
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		_, err := dstColl.InsertMany(ctx, batch)
		if err != nil {
			logger.Error("Failed to insert final batch", "db", task.DstDB, "coll", task.Coll, "err", err)
		} else {
			totalCopied += len(batch)
			progress.Advance(workerID, int64(len(batch)))
		}
	}

	if err := cursor.Err(); err != nil {
		logger.Error("Cursor error after reading collection", "db", task.SrcDB, "coll", task.Coll, "err", err)
	}
	_ = cursor.Close(ctx)

	indexCount, err := copyIndexes(ctx, srcColl, dstColl)
	if err != nil {
		logger.Error("Failed to copy indexes", "db", task.SrcDB, "coll", task.Coll, "err", err)
		return
	}

	logger.Info("Copied collection", "db", task.SrcDB, "coll", task.Coll, "count", totalCopied)
	if indexCount > 0 {
		logger.Info("Copied indexes", "db", task.SrcDB, "coll", task.Coll, "index_count", indexCount)
	}
}

func copyIndexes(ctx context.Context, srcColl, dstColl *mongo.Collection) (int, error) {
	indexSpecs, err := srcColl.Indexes().ListSpecifications(ctx)
	if err != nil {
		return 0, err
	}

	models := make([]mongo.IndexModel, 0, len(indexSpecs))
	for _, spec := range indexSpecs {
		if spec == nil || spec.Name == "_id_" {
			continue
		}

		var keys bson.D
		if err := bson.Unmarshal(spec.KeysDocument, &keys); err != nil {
			return 0, err
		}

		indexOpts := options.Index().SetName(spec.Name)
		if spec.Unique != nil {
			indexOpts.SetUnique(*spec.Unique)
		}
		if spec.Sparse != nil {
			indexOpts.SetSparse(*spec.Sparse)
		}
		if spec.ExpireAfterSeconds != nil {
			indexOpts.SetExpireAfterSeconds(*spec.ExpireAfterSeconds)
		}

		models = append(models, mongo.IndexModel{
			Keys:    keys,
			Options: indexOpts,
		})
	}

	if len(models) == 0 {
		return 0, nil
	}

	if _, err := dstColl.Indexes().CreateMany(ctx, models); err != nil {
		return 0, err
	}

	return len(models), nil
}
