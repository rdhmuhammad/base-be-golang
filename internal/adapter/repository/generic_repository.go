package repository

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type GenericRepository[T interface{}] struct {
	db *gorm.DB
}

func (repo *GenericRepository[T]) SetupConnection(db *gorm.DB) {
	repo.db = db
}

func NewGenericeRepoPointr[T any](db *gorm.DB) *GenericRepository[T] {
	return &GenericRepository[T]{
		db: db,
	}
}

func NewGenericeRepo[T any](db *gorm.DB) GenericRepository[T] {
	return GenericRepository[T]{
		db: db,
	}
}

var (
	getColName = func(col string) (string, string) {
		var colName, tableName string
		_, err := fmt.Sscanf(col, "%s.%s", &colName, &tableName)
		if err != nil {
			log.Println("invalid coloumn name, format is table.column")
			return col, ""
		}

		return colName, tableName
	}

	ExpressionSearch = func(val interface{}, colName ...string) clause.Expression {
		var exps = make([]clause.Expression, len(colName))
		for i, coln := range colName {
			col, table := getColName(coln)
			exps[i] = clause.Like{
				Column: clause.Column{Name: col, Table: table},
				Value:  val,
			}
		}
		return clause.Or(exps...)
	}

	ExpressionEqual = func(val interface{}, col string) clause.Expression {
		colName, tableName := getColName(col)
		return clause.Eq{
			Column: clause.Column{Name: colName, Table: tableName},
			Value:  val,
		}

	}
	ExpressionDateRange = func(start time.Time, end time.Time, col string, table string) clause.Expression {
		return clause.And(
			clause.Gte{
				Column: clause.Column{Name: col},
				Value:  start,
			},
			clause.Lte{
				Column: clause.Column{Name: col},
				Value:  end,
			},
		)
	}
)

type PaginationQuery struct {
	PerPage int
	Page    int
}

type GenericRepositoryInterface[T any] interface {
	FindAllByExpression(
		ctx context.Context,
		expression []clause.Expression,
	) ([]T, error)
	FindAll(ctx context.Context) ([]T, error)
	FindAllByExpressionAndJoin(
		ctx context.Context,
		cond []clause.Expression,
		join []string,
		preload []string,
	) ([]T, error)
	FindAllByScopeAndJoin(
		ctx context.Context,
		scope []func(db *gorm.DB) *gorm.DB,
		join []string,
		preload []string,
	) ([]T, error)
	Update(ctx context.Context, data T) error
	UpdateSelectedCols(ctx context.Context, data T, columns ...string) error
	BulkStore(ctx context.Context, data []T) ([]T, error)
	Store(ctx context.Context, data T) (T, error)
	Delete(ctx context.Context, data T) error
	FindOneByID(ctx context.Context, id interface{}) (T, error)
	FindOneByExpressionAndJoin(
		ctx context.Context,
		cond []clause.Expression,
		joins []string,
		preload []string,
	) (T, error)
	FindOneByExpression(
		ctx context.Context,
		cond []clause.Expression,
	) (T, error)
	FindAllByExpressionPaginate(
		ctx context.Context,
		paginate PaginationQuery,
		cond []clause.Expression,
	) ([]T, int, error)
	BulkUpdateSelectedColumn(ctx context.Context, children []T, fields ...string) error
}

func (repo GenericRepository[T]) BulkUpdateSelectedColumn(ctx context.Context, children []T, fields ...string) error {
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, child := range children {
			err := tx.Select(fields).
				Updates(&child).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (repo GenericRepository[T]) Update(ctx context.Context, data T) error {
	return repo.db.WithContext(ctx).Updates(&data).Error
}

func (repo GenericRepository[T]) UpdateSelectedCols(ctx context.Context, data T, columns ...string) error {
	return repo.db.WithContext(ctx).
		Select(columns).
		Updates(&data).Error
}

func (repo GenericRepository[T]) BulkStore(ctx context.Context, data []T) ([]T, error) {
	err := repo.db.WithContext(ctx).
		Create(&data).Error

	return data, err
}

func (repo GenericRepository[T]) Store(ctx context.Context, data T) (T, error) {
	err := repo.db.WithContext(ctx).
		Create(&data).Error

	return data, err
}

func (repo GenericRepository[T]) Delete(ctx context.Context, data T) error {
	return repo.db.WithContext(ctx).Delete(&data).Error
}

func (repo GenericRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	var data []T
	err := repo.db.WithContext(ctx).
		Find(&data).Error

	return data, err
}

func (repo GenericRepository[T]) FindAllByExpression(
	ctx context.Context,
	expression []clause.Expression,
) ([]T, error) {
	var result []T
	err := repo.db.WithContext(ctx).
		Clauses(clause.Where{Exprs: expression}).
		Find(&result).Error

	return result, err
}

func (repo GenericRepository[T]) FindAllByScopeAndJoin(
	ctx context.Context,
	scope []func(db *gorm.DB) *gorm.DB,
	join []string,
	preload []string,
) ([]T, error) {
	var result []T
	db := repo.db.WithContext(ctx).
		Model(&result)

	for _, sc := range scope {
		db = db.Scopes(sc)
	}

	for _, pr := range preload {
		db = db.Preload(pr)
	}

	for _, j := range join {
		db = db.Joins(j)
	}

	err := db.First(&result).Error
	return result, err
}

func (repo GenericRepository[T]) FindOneByExpressionAndJoin(
	ctx context.Context,
	cond []clause.Expression,
	joins []string,
	preload []string,
) (T, error) {
	var result T
	db := repo.db.WithContext(ctx).
		Model(&result).
		Clauses(clause.Where{Exprs: cond})

	for _, pr := range preload {
		db = db.Preload(pr)
	}

	for _, join := range joins {
		db = db.Joins(join)
	}

	err := db.First(&result).Error
	return result, err
}

func (repo GenericRepository[T]) FindOneByID(ctx context.Context, id interface{}) (T, error) {
	var data T
	err := repo.db.WithContext(ctx).
		First(&data, "id = ?", id).Error

	return data, err
}

func (repo GenericRepository[T]) FindOneByExpression(
	ctx context.Context,
	cond []clause.Expression,
) (T, error) {
	var result T
	db := repo.db.WithContext(ctx).
		Model(&result).
		Clauses(clause.Where{Exprs: cond})

	err := db.First(&result).Error
	return result, err
}

func (repo GenericRepository[T]) FindAllByExpressionAndJoin(
	ctx context.Context,
	cond []clause.Expression,
	join []string,
	preload []string,
) ([]T, error) {
	var result []T
	db := repo.db.WithContext(ctx).
		Model(&result).Clauses(clause.Where{Exprs: cond})

	for _, j := range join {
		db = db.Joins(j)
	}

	for _, s := range preload {
		db = db.Preload(s)
	}

	err := db.Find(&result).Error
	return result, err
}

func (repo GenericRepository[T]) FindAllByExpressionPaginate(
	ctx context.Context,
	paginate PaginationQuery,
	cond []clause.Expression,
) ([]T, int, error) {
	var result []T
	var total int64
	db := repo.db.WithContext(ctx).
		Model(&result).
		Clauses(clause.Where{Exprs: cond})

	errCount := db.Count(&total).Error
	if errCount != nil {
		return nil, 0, errCount
	}

	offset := paginate.PerPage * (paginate.Page - 1)
	limit := paginate.PerPage

	err := db.Find(&result).
		Limit(limit).
		Offset(offset).Error

	return result, int(total), err
}
