package utils

import (
	"context"
	"database/sql"

	"tutuplapak-go/repository"
)

func GetFileInfo(queries *repository.Queries, ctx context.Context, fileID sql.NullInt32) (string, string, error) {
	if !fileID.Valid || fileID.Int32 <= 0 {
		return "", "", nil
	}

	file, err := queries.GetFileByID(ctx, fileID.Int32)
	if err != nil {
		return "", "", err
	}

	fileURI := file.FileUri
	fileThumbnailURI := ""
	if file.FileThumnailUri.Valid {
		fileThumbnailURI = file.FileThumnailUri.String
	}

	return fileURI, fileThumbnailURI, nil
}
