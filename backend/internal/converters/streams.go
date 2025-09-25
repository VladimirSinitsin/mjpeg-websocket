package converters

import (
	"fmt"
	v1 "stream-server/api/v1"
	"stream-server/internal/data/repo"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToApiStreamResponseList(in []repo.ListStreamsRow) (res []*v1.Stream) {
	for _, row := range in {
		item := &v1.Stream{
			Id:              row.ID.String(),
			Title:           row.Title,
			Description:     row.Description,
			FrameIntervalMs: row.FrameIntervalMs,
			CreatedAt:       timestamppb.New(row.CreatedAt.Time),
			UpdatedAt:       timestamppb.New(row.UpdatedAt.Time),
			FrameCount:      row.FrameCount,
		}
		res = append(res, item)
	}

	return res
}

func ToApiStreamUpdateResult(in repo.UpdateStreamRow) *v1.Stream {
	return &v1.Stream{
		Id:              in.ID.String(),
		Title:           in.Title,
		Description:     in.Description,
		FrameIntervalMs: in.FrameIntervalMs,
		CreatedAt:       timestamppb.New(in.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(in.UpdatedAt.Time),
		FrameCount:      in.FrameCount,
	}
}

func ToApiStreamResponse(row repo.GetStreamRow) *v1.Stream {
	return &v1.Stream{
		Id:              row.ID.String(),
		Title:           row.Title,
		Description:     row.Description,
		FrameIntervalMs: row.FrameIntervalMs,
		CreatedAt:       timestamppb.New(row.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(row.UpdatedAt.Time),
		FrameCount:      row.FrameCount,
	}
}

func ToDbUpdateStreamParams(in *v1.UpdateStreamRequest) (res repo.UpdateStreamParams, err error) {
	uuid, err := StringToPgUUID(in.Id)
	if err != nil {
		return res, fmt.Errorf("error converting uuid: %w", err)
	}

	return repo.UpdateStreamParams{
		ID:              uuid,
		Title:           in.Title,
		Description:     in.Description,
		FrameIntervalMs: in.FrameIntervalMs,
	}, nil
}
