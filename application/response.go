package application

import "github.com/allentom/haruka"

func SendSuccessResponse(ctx *haruka.Context) {
	ctx.Writer.WriteHeader(200)
	ctx.JSON(map[string]interface{}{
		"success":true,
	})
}