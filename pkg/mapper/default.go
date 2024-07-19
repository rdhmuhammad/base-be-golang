//go:generate mockery --all --inpackage --case snake

package mapper

import (
	"base-be-golang/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type Mapper struct {
}

type MapperUtility interface {
	NewResponse(c *gin.Context, res *dto.Response, err error)
	ReplaceLabelErr(template error, params ...string) error
	ErrorIs(template error, targer error) bool
	TranslateSQLErr(mySqlErr *mysql.MySQLError, methodName string) error
	IsInvalidDataError(err error) bool
	IsAccessControlError(err error) bool
	CompareSliceOfErr(errs []error, target error) bool
	ParseServiceDurationFormat(d string) (string, error)
	SortingByStructField(vals interface{}, fieldName string, sorting SortingDirection) interface{}
	UniqueByStructField(vals interface{}, fieldName string) interface{}
}
