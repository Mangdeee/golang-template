package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func StatusInList(status int, statusList []int) bool {
	for _, i := range statusList {
		if i == status {
			return true
		}
	}
	return false
}

func DbTransactionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		txHandle := db.Begin()
		log.Print("begin db transaction")

		defer func() {
			if r := recover(); r != nil {
				txHandle.Rollback()
			}
		}()

		c.Set("db_trx", txHandle)
		c.Next()

		if StatusInList(c.Writer.Status(), []int{http.StatusOK, http.StatusCreated}) {
			err := txHandle.Commit().Error
			if err != nil {
				log.Print("trx commit error: ", err)
			}
		} else {
			log.Print("rolling back transaction due to status code: ", c.Writer.Status())
			txHandle.Rollback()
		}
	}
}
