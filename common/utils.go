package common

import (
	"github.com/sirupsen/logrus"
)

func HandleError(err error, errType ErrorType) {
	if err != nil {
		switch errType {
		case ErrorFatal:
			logrus.Fatal(err)
		case ErrorError:
			logrus.Error(err)
		}
	}
}
