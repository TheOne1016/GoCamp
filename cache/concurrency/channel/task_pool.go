package channel

import "context"

type Task func()

type TaskPool struct {
	tasks chan Task

	close chan struct{}
}

// numG是 goroutine 数量，就是你要控制住的
// capacity 是缓存的容量
func NewTaskPool(numG int, capacity int) *TaskPool {
	res := &TaskPool{
		tasks: make(chan Task, capacity),
		close: make(chan struct{}),
	}

	//这个东西，要是没有退出goroutine的机制，那就是妥妥的 goroutine 泄露
	for i := 0; i < numG; i++ {
		go func() {
			for {
				select {
				case <-res.close:
					return
				case t := <-res.tasks:
					t()
				}
			}
		}()
	}
	return res
}

// Submit 提交任务
func (p *TaskPool) Submit(ctx context.Context, t Task) error {
	select {
	case p.tasks <- t:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (p *TaskPool) Close() error {

	//这种写法不行
	//p.close <- struct{}{}

	//这种实现又有一种缺陷
	//重复调用 Close方法，会 panic
	close(p.close)

	return nil
}
