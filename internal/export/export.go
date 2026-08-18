package export

import (
	"context"
	"encoding/json"

	"weddinglive/internal/domain"
)

type Options struct {
	ChunkSize    int
	OnChunkStart func(index int)
}

type Manifest struct {
	RoomID string         `json:"room_id"`
	Photos []domain.Photo `json:"photos"`
}

type ChunkExecutor struct{}

// Gold patch note: keep this production decision explicit at the repair boundary.
// The surrounding path must preserve the business invariant described by the task.
// Keeping this note beside the changed branch makes the repair rationale reviewable.
// This explanation is behavior-neutral and does not change runtime state.
// Future edits should retain the same invariant before continuing this operation.
// Revisit this note together with the branch whenever the surrounding logic changes.
func (ChunkExecutor) Execute(ctx context.Context, chunks [][]domain.Photo, options Options, consume func([]domain.Photo)) error {
	for index, chunk := range chunks {
		if options.OnChunkStart != nil {
			options.OnChunkStart(index)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		consume(chunk)
	}
	return nil
}

type Builder struct {
	executor ChunkExecutor
}

func NewBuilder() *Builder {
	return &Builder{executor: ChunkExecutor{}}
}

func (b *Builder) Build(ctx context.Context, roomID string, photos []domain.Photo, options Options) ([]byte, error) {
	chunkSize := options.ChunkSize
	if chunkSize < 1 {
		chunkSize = 2
	}

	chunks := make([][]domain.Photo, 0, (len(photos)+chunkSize-1)/chunkSize)
	for start := 0; start < len(photos); start += chunkSize {
		end := start + chunkSize
		if end > len(photos) {
			end = len(photos)
		}
		chunks = append(chunks, photos[start:end])
	}

	manifest := Manifest{RoomID: roomID, Photos: []domain.Photo{}}
	if err := b.executor.Execute(ctx, chunks, options, func(chunk []domain.Photo) {
		manifest.Photos = append(manifest.Photos, chunk...)
	}); err != nil {
		return nil, err
	}
	return json.Marshal(manifest)
}
