package async

// import (
// 	"GoCamp/webook/internal/repository"
// 	"GoCamp/webook/internal/service/sms"
// 	"context"
// )

// /*
// 	此源代码是 第六次作业

// */

// type SMSService struct {
// 	svc  sms.Service
// 	repo repository.SMSAsyncReqRepository
// }

// func NewSMSService() *SMSService {
// 	return &SMSService{}
// }

// func (s *SMSService) StartAsyna() {
// 	go func() {
// 		reqs := repo.Find没发出去的请求()
// 		for _, req := range reqs {
// 			//在这里发送，并控制重试
// 		}
// 	}()
// }

// func (s *SMSService) Send(ctx context.Context, biz string,
// 	args []string, numbers ...string) error {
// 	//首先是正常路径
// 	err := s.svc.Send(ctx, biz, args, numbers...)
// 	if err != nil {
// 		// 判断是不是崩溃

// 		if 崩溃了 {
// 			s.repo.Store()
// 		}
// 	}
// 	return
// }
