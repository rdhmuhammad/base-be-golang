package mapper

import (
	"base-be-golang/internal/dto"
	localerror "base-be-golang/internal/localerror"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"regexp"
	"strings"
)

// TODO: validation move to struct tags
func (m Mapper) TranslateSQLErr(mySqlErr *mysql.MySQLError, methodName string) error {
	switch mySqlErr.Number {
	case DuplicateEntryCode:
		re := regexp.MustCompile(`for key '([^']+)`)
		msgs := re.FindStringSubmatch(mySqlErr.Message)
		if len(msgs) > 1 {
			switch msgs[1] {
			//case "kassir_users.uk_kassir_users_username":
			//	return localerror.InvalidDataError{Msg: ErrUserNameIsUsed.Error()}
			default:
				return mySqlErr
			}
		}
		return mySqlErr

	case DataTruncateCode:
		re := regexp.MustCompile(`for column '([^']+)`)
		msgs := re.FindStringSubmatch(mySqlErr.Message)
		if len(msgs) > 1 {
			switch msgs[1] {
			case "gender":
				//return localerror.InvalidDataError{
				//	Msg: ErrInvalidGender.Error(),
				//}
			default:
				return mySqlErr
			}
		}
		return fmt.Errorf("%s: %w", methodName, mySqlErr)

	case ForeignConstrainFailCode:
		re := regexp.MustCompile("CONSTRAINT `([^`]+)")
		msgs := re.FindStringSubmatch(mySqlErr.Message)
		if len(msgs) > 1 {
			switch msgs[1] {
			//case "fk_kassir_branches":
			//	return localerror.InvalidDataError{
			//		Msg: ErrBranchNotFound.Error(),
			//	}
			default:
				return mySqlErr
			}
		}

	default:
		return mySqlErr
	}

	return mySqlErr

}

func (m Mapper) NewResponse(c *gin.Context, res *dto.Response, err error) {
	if err != nil {
		if m.IsInvalidDataError(err) {
			c.JSON(http.StatusBadRequest, dto.DefaultErrorInvalidDataWithMessage(err.Error()))
			return
		}
		if m.IsAccessControlError(err) {
			c.JSON(http.StatusUnauthorized, dto.DefaultErrorInvalidDataWithMessage(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.DefaultErrorResponseWithMessage(err.Error()))
	}
	if res != nil {
		c.JSON(http.StatusOK, res)
		return
	}

	c.Status(http.StatusOK)
}

func (m Mapper) IsInvalidDataError(err error) bool {
	var invalidDataError localerror.InvalidDataError
	if errors.As(err, &invalidDataError) {
		return true
	}
	return false
}

func (m Mapper) IsAccessControlError(err error) bool {
	var invalidDataError localerror.AccessControlError
	if errors.As(err, &invalidDataError) {
		return true
	}
	return false
}

func (m Mapper) CompareSliceOfErr(errs []error, target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
		if m.ErrorIs(err, target) {
			return true
		}
	}

	return false
}

func (m Mapper) ErrorIs(template error, targer error) bool {
	re := regexp.MustCompile(`\{[0-9]+}`)
	pattern := re.ReplaceAllString(template.Error(), ".+")

	match, err := regexp.MatchString(pattern+"$", targer.Error())
	if err != nil {

		return false
	}

	if match {
		return true
	}

	return false
}

func (m Mapper) ReplaceLabelErr(template error, params ...string) error {
	customeErr := template.Error()
	for i, param := range params {
		customeErr = strings.Replace(
			customeErr,
			fmt.Sprintf("{%d}", i),
			param,
			-1,
		)
	}

	return fmt.Errorf(customeErr)
}
