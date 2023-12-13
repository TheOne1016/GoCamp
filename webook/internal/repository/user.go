package repository

import (
	"GoCamp/webook/internal/domain"
	"GoCamp/webook/internal/repository/cache"
	"GoCamp/webook/internal/repository/dao"
	"context"
	"database/sql"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDao
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDao, c *cache.UserCache) *UserRepository {
	return &UserRepository{dao: dao,
		cache: c,
	}
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *UserRepository) SaveEditInfo(ctx context.Context, u domain.User) error {
	return r.dao.Update(ctx, r.domainToEntity(u))
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserRepository) FindById(ctx context.Context, Id int64) (domain.User, error) {
	//先从cache里面找
	//再从dao里面找
	//找到了回写cache
	u, err := r.cache.Get(ctx, Id)

	//必然有数据
	if err == nil {
		return u, nil
	}

	// //没这个数据
	// if err ==cache.ErrKeyNotExist {
	// 	//去数据库里面加载
	// }

	ue, err := r.dao.FindById(ctx, Id)
	if err != nil {
		return domain.User{}, err
	}

	u = r.entityToDomain(ue)

	go func() {
		err = r.cache.Set(ctx, u)
		if err != nil {
			//打日志，做监控
		}

	}()

	return u, err

	//这里怎么办，要不要从数据库中加载
	//选加载————做好兜底，万一Redis真的崩了，要保护住数据库(数据库限流)
	//选不加载————用户体验差一点

}

func (r *UserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password:  u.Password,
		Ctime:     u.Ctime.Unix(),
		NickName:  u.NickName,
		Birthday:  u.Birthday,
		BirefInfo: u.BirefInfo,
	}
}

func (r *UserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:        u.Id,
		Email:     u.Email.String,
		Password:  u.Password,
		Phone:     u.Phone.String,
		Ctime:     time.UnixMilli(u.Ctime),
		NickName:  u.NickName,
		Birthday:  u.Birthday,
		BirefInfo: u.BirefInfo,
	}
}
