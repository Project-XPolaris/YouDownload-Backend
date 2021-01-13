package application

import "github.com/allentom/haruka"

func Abort500Error(err error,ctx *haruka.Context) {
	AppLogger.Error(err)
	ctx.Writer.WriteHeader(500)
	ctx.JSON(map[string]interface{}{
		"success":false,
		"err":err.Error(),
	})
}