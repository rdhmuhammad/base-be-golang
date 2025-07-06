//go:generate mockery --all --inpackage --case snake

package environment

import (
	"log"
	"os"
	"strconv"
)

type Environment interface {
	Get(key string) string
	GetUint(key string, defaultValue uint) uint
	GetBranchID() uint
}

func NewEnvironment() Environment {
	return environment{}
}

type environment struct{}

func (e environment) GetBranchID() uint {
	return e.GetUint("BRANCH_ID", 1)
}

func (environment) Get(key string) string {
	return os.Getenv(key)
}

func (e environment) GetUint(key string, defaultValue uint) uint {
	str := os.Getenv(key)
	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		log.Println("featureflag.environment.GetUint:", err)
		value = uint64(defaultValue)
	}

	return uint(value)
}
