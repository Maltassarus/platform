package scheduler

import (
	"context"
	"log"
	"platform/internal/service"
	"platform/internal/worker"
	"time"
)

type Scheduler struct {
	postService   *service.PostService
	checkInterval time.Duration
	workerPool    *worker.WorkerPool
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewScheduler(postService *service.PostService, checkInterval time.Duration, poolSize int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	pool := worker.NewWorkerPool(ctx, poolSize, func(job worker.Job) {
		if err := postService.PublishPost(job.PostID); err != nil {
			log.Printf("Failed to publish post %d: %v", job.PostID, err)
		} else {
			log.Printf("Post %d published successfully by scheduler", job.PostID)
		}
	})

	return &Scheduler{
		postService:   postService,
		checkInterval: checkInterval,
		workerPool:    pool,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(s.checkInterval)
	go func() {
		log.Printf("Scheduler started: checking every %v", s.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkAndPublish()
			case <-s.ctx.Done():
				log.Println("Scheduler stopped")
				return
			}
		}
	}()
}

func (s *Scheduler) checkAndPublish() {
	posts, err := s.postService.GetPendingPosts()
	if err != nil {
		log.Printf("Error fetching pending posts: %v", err)
		return
	}
	if len(posts) == 0 {
		return
	}
	log.Printf("Found %d posts ready for publication", len(posts))
	for _, post := range posts {
		s.workerPool.Submit(worker.Job{PostID: post.ID})
	}
}

func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.cancel()
	s.workerPool.Stop()
	log.Println("Scheduler stopped")
}
