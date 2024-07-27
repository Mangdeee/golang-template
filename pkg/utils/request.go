package utils

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gobeam/stringy"

	"github.com/felixlambertv/go-cleanplate/internal/model"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/gin-gonic/gin"
)

var (
	suffixes [5]string
)

func CheckWhitelistUrl(url string) bool {
	splittedUrl := strings.Split(url, "api/v1/")
	whitelistedUrl := map[string]bool{
		"users/me":           true,
		"auth/refresh-token": true,
	}

	return whitelistedUrl[splittedUrl[1]]
}

func GeneratePaginationFromRequest(ctx *gin.Context, dbModel interface{}) model.Pagination {
	// Initializing default
	//	var mode string
	limit := 10
	page := 1
	search := ""
	level := ""
	sort := "id"
	direction := "asc"
	var createdFrom time.Time
	var createdTo time.Time

	query := ctx.Request.URL.Query()
	for key, value := range query {
		queryValue := value[len(value)-1]
		switch key {
		case "limit":
			limit, _ = strconv.Atoi(queryValue)
			break
		case "page":
			page, _ = strconv.Atoi(queryValue)
			break
		case "q":
			search = queryValue
			break
		case "sort":
			if queryValue != "" {
				str := stringy.New(queryValue)
				snakeStr := str.SnakeCase("?", "")
				sort = snakeStr.ToLower()
			}
			break
		case "direction":
			queryValue = strings.ToLower(queryValue)
			if queryValue == "asc" || queryValue == "desc" {
				direction = queryValue
			}

			break
		case "level":
			queryValue = strings.ToLower(queryValue)
			if queryValue == "beginner" || queryValue == "intermediate" || queryValue == "advanced" {
				level = queryValue
			}
			break
		case "createdFrom":
			queryValue = strings.ToLower(queryValue)
			if queryValue != "" {
				date, err := time.Parse(consttype.DATEFORMAT, queryValue)
				if err == nil {
					createdFrom = date
				}
			}
			break
		case "createdTo":
			queryValue = strings.ToLower(queryValue)
			if queryValue != "" {
				date, err := time.Parse(consttype.DATEFORMAT, queryValue)
				if err == nil {
					createdTo = date
				}
			}
			break
		}
	}

	return model.Pagination{
		Limit:     limit,
		Page:      page,
		Sort:      sort,
		Direction: direction,
		Search:    search,
		Filter: model.Filter{
			Level:       level,
			CreatedFrom: createdFrom,
			CreatedTo:   createdTo,
		},
	}
}

func GetExtensionType(ext string) string {
	images := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	video := map[string]bool{".mp4": true}
	sound := map[string]bool{".mp3": true}

	filetype := "others"

	if images[ext] {
		filetype = "image"
	} else if video[ext] {
		filetype = "video"
	} else if sound[ext] {
		filetype = "sound"
	}

	return filetype
}

func ValidateExtension(ext string) error {
	valid := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".mp4": true, ".mp3": true}

	if valid[ext] {
		return nil
	} else {
		keys := reflect.ValueOf(valid).MapKeys()
		return fmt.Errorf("only allowed file extension : %s", keys)
	}
}

func GetReadableFileSize(size float64, ext string) error {
	suffixes[0] = "B"
	suffixes[1] = "KB"
	suffixes[2] = "MB"
	suffixes[3] = "GB"
	suffixes[4] = "TB"
	maxImageFile := 0.6
	maxImageFileSuffix := "MB"
	maxVideoSize := 30
	maxVideoSizeSuffix := "MB"
	maxSoundSize := 10
	maxSoundSizeSuffix := "MB"

	base := math.Log(size) / math.Log(1024)
	getSize := Round(math.Pow(1024, base-math.Floor(base)), .5, 2)
	getSuffix := suffixes[int(math.Floor(base))]
	if getSuffix == "KB" {
		getSize = math.Ceil((getSize/1000)*100) / 100
		getSuffix = "MB"
	}

	if maxImageFileSuffix == getSuffix || maxVideoSizeSuffix == getSuffix {
		switch ext {
		case "image":
			if float64(getSize) > maxImageFile || getSuffix == "GB" || getSuffix == "TB" {
				return fmt.Errorf("image size is too big. Maximum size is %f %s", float64(maxImageFile), maxImageFileSuffix)
			}
			return nil
		case "video":
			if int(getSize) > maxVideoSize || getSuffix == "GB" || getSuffix == "TB" {
				return fmt.Errorf("video size is too big. Maximum size is %d %s", maxVideoSize, maxVideoSizeSuffix)
			}
			return nil
		case "sound":
			if int(getSize) > maxSoundSize || getSuffix == "GB" || getSuffix == "TB" {
				return fmt.Errorf("sound size is too big. Maximum size is %d %s", maxSoundSize, maxSoundSizeSuffix)
			}
			return nil
		}
	}
	return nil
}
