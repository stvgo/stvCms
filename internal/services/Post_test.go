package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"stvCms/internal/mocks"
	"stvCms/internal/models"
	"stvCms/internal/rest/request"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func newTestService(ctrl *gomock.Controller) (*postService, *mocks.MockIPostRepository, *mocks.MockIRedisClient, *mocks.MockIR2Client, *mocks.MockIOpenRouterClient) {
	repo := mocks.NewMockIPostRepository(ctrl)
	redis := mocks.NewMockIRedisClient(ctrl)
	r2 := mocks.NewMockIR2Client(ctrl)
	ai := mocks.NewMockIOpenRouterClient(ctrl)

	svc := &postService{
		repository:       repo,
		ctx:              context.Background(),
		redisClient:      redis,
		openRouterClient: ai,
		r2:               r2,
	}
	return svc, repo, redis, r2, ai
}

// --- CreatePost ---

func TestCreatePost(t *testing.T) {
	tests := []struct {
		name        string
		req         request.CreatePostRequest
		setupMocks  func(*mocks.MockIPostRepository, *mocks.MockIRedisClient)
		wantErr     bool
		wantContain string
	}{
		{
			name: "éxito con content blocks",
			req: request.CreatePostRequest{
				Title:  "Test Post",
				UserID: "user1",
				ContentBlocks: []request.ContentBlock{
					{Type: "text", Order: 1, Content: "Hello"},
				},
			},
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				repo.EXPECT().CreatePost(gomock.Any()).Return("Post creado", nil)
				redis.EXPECT().Del(gomock.Any(), "posts:all").Return(nil)
			},
			wantErr:     false,
			wantContain: "Post creado",
		},
		{
			name: "repo falla",
			req:  request.CreatePostRequest{Title: "Fail"},
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				repo.EXPECT().CreatePost(gomock.Any()).Return("", errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, repo, redis, _, _ := newTestService(ctrl)
			tt.setupMocks(repo, redis)

			result, err := svc.CreatePost(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, result, tt.wantContain)
			}
		})
	}
}

// --- GetPosts ---

func TestGetPosts(t *testing.T) {
	post := models.Post{
		Title:  "Cached",
		UserID: "u1",
		ContentBlocks: []models.ContentBlock{
			{Type: "text", Order: 1, Content: "body"},
		},
	}
	cachedData, _ := json.Marshal([]models.Post{post})

	tests := []struct {
		name       string
		setupMocks func(*mocks.MockIPostRepository, *mocks.MockIRedisClient)
		wantLen    int
		wantErr    bool
	}{
		{
			name: "cache hit",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				redis.EXPECT().Get(gomock.Any(), "posts:all").Return(string(cachedData), nil)
			},
			wantLen: 1,
		},
		{
			name: "cache miss, DB ok",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				redis.EXPECT().Get(gomock.Any(), "posts:all").Return("", errors.New("cache miss"))
				repo.EXPECT().GetPosts().Return([]models.Post{post}, nil)
				redis.EXPECT().Set(gomock.Any(), "posts:all", gomock.Any(), 24*time.Hour).Return(nil)
			},
			wantLen: 1,
		},
		{
			name: "cache miss, DB falla",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				redis.EXPECT().Get(gomock.Any(), "posts:all").Return("", errors.New("cache miss"))
				repo.EXPECT().GetPosts().Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "cache miss, DB vacío",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				redis.EXPECT().Get(gomock.Any(), "posts:all").Return("", errors.New("cache miss"))
				repo.EXPECT().GetPosts().Return([]models.Post{}, nil)
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, repo, redis, _, _ := newTestService(ctrl)
			tt.setupMocks(repo, redis)

			result, err := svc.GetPosts()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}
		})
	}
}

// --- GetPostById ---

func TestGetPostById(t *testing.T) {
	post := models.Post{
		Title:  "Detail",
		UserID: "u1",
		ContentBlocks: []models.ContentBlock{
			{Type: "code", Order: 1, Content: "fmt.Println()", Language: "go"},
		},
	}

	tests := []struct {
		name       string
		id         int
		setupMocks func(*mocks.MockIPostRepository)
		wantErr    bool
	}{
		{
			name: "encontrado",
			id:   1,
			setupMocks: func(repo *mocks.MockIPostRepository) {
				repo.EXPECT().GetPostById(uint(1)).Return(post, nil)
			},
		},
		{
			name: "no encontrado",
			id:   99,
			setupMocks: func(repo *mocks.MockIPostRepository) {
				repo.EXPECT().GetPostById(uint(99)).Return(models.Post{}, gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, repo, _, _, _ := newTestService(ctrl)
			tt.setupMocks(repo)

			result, err := svc.GetPostById(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, post.Title, result.Title)
				assert.Len(t, result.ContentBlocks, 1)
			}
		})
	}
}

// --- GetPostByFilter ---

func TestGetPostByFilter(t *testing.T) {
	posts := []models.Post{
		{Title: "Go tutorial", UserID: "u1"},
		{Title: "Python guide", UserID: "u2"},
	}

	tests := []struct {
		name       string
		filter     string
		setupMocks func(*mocks.MockIPostRepository)
		wantLen    int
		wantErr    bool
	}{
		{
			name:   "con resultados",
			filter: "Go",
			setupMocks: func(repo *mocks.MockIPostRepository) {
				repo.EXPECT().GetPostsByFilter("Go").Return(posts[:1], nil)
			},
			wantLen: 1,
		},
		{
			name:   "sin resultados",
			filter: "xxx",
			setupMocks: func(repo *mocks.MockIPostRepository) {
				repo.EXPECT().GetPostsByFilter("xxx").Return([]models.Post{}, nil)
			},
			wantLen: 0,
		},
		{
			name:   "repo error",
			filter: "bad",
			setupMocks: func(repo *mocks.MockIPostRepository) {
				repo.EXPECT().GetPostsByFilter("bad").Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, repo, _, _, _ := newTestService(ctrl)
			tt.setupMocks(repo)

			result, err := svc.GetPostByFilter(tt.filter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}
		})
	}
}

// --- UpdatePost ---

func TestUpdatePost(t *testing.T) {
	req := request.UpdatePostRequest{
		Id:    1,
		Title: "Updated",
		ContentBlocks: []request.ContentBlock{
			{Type: "text", Order: 1, Content: "new content"},
		},
	}

	tests := []struct {
		name       string
		setupMocks func(*mocks.MockIPostRepository, *mocks.MockIRedisClient)
		wantErr    bool
	}{
		{
			name: "éxito",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				repo.EXPECT().ExistsPost(1).Return(true)
				repo.EXPECT().UpdatePost(uint(1), gomock.Any()).Return("Post actualizado", nil)
				redis.EXPECT().Del(gomock.Any(), "posts:all").Return(nil)
			},
		},
		{
			name: "post no existe",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				repo.EXPECT().ExistsPost(1).Return(false)
			},
			wantErr: true,
		},
		{
			name: "repo update falla",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				repo.EXPECT().ExistsPost(1).Return(true)
				repo.EXPECT().UpdatePost(uint(1), gomock.Any()).Return("", errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, repo, redis, _, _ := newTestService(ctrl)
			tt.setupMocks(repo, redis)

			_, err := svc.UpdatePost(req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- DeletePostById ---

func TestDeletePostById(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setupMocks func(*mocks.MockIPostRepository, *mocks.MockIRedisClient)
		wantErr    bool
	}{
		{
			name: "éxito",
			id:   "5",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				repo.EXPECT().DeletePostById(5).Return(true)
				redis.EXPECT().Del(gomock.Any(), "posts:all").Return(nil)
			},
		},
		{
			name: "no encontrado",
			id:   "99",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				repo.EXPECT().DeletePostById(99).Return(false)
			},
			wantErr: true,
		},
		{
			name: "ID inválido (se convierte en 0, no encontrado)",
			id:   "abc",
			setupMocks: func(repo *mocks.MockIPostRepository, redis *mocks.MockIRedisClient) {
				repo.EXPECT().DeletePostById(0).Return(false)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, repo, redis, _, _ := newTestService(ctrl)
			tt.setupMocks(repo, redis)

			_, err := svc.DeletePostById(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- GetImage ---

func TestGetImage(t *testing.T) {
	imageData := []byte("fake image bytes")

	tests := []struct {
		name       string
		filename   string
		setupMocks func(*mocks.MockIR2Client)
		wantErr    bool
	}{
		{
			name:     "R2 ok",
			filename: "photo.jpg",
			setupMocks: func(r2 *mocks.MockIR2Client) {
				r2.EXPECT().GetObject(gomock.Any(), gomock.Any()).Return(
					&s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(imageData))},
					nil,
				)
			},
		},
		{
			name:     "R2 falla",
			filename: "missing.jpg",
			setupMocks: func(r2 *mocks.MockIR2Client) {
				r2.EXPECT().GetObject(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, _, _, r2, _ := newTestService(ctrl)
			tt.setupMocks(r2)

			data, err := svc.GetImage(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, imageData, data)
			}
		})
	}
}

// --- AutoCompleteAI ---

func TestAutoCompleteAI(t *testing.T) {
	tests := []struct {
		name       string
		req        request.AI
		setupMocks func(*mocks.MockIOpenRouterClient)
		wantErr    bool
		wantResult string
	}{
		{
			name: "texto AI ok",
			req:  request.AI{TextAI: "Write about Go"},
			setupMocks: func(ai *mocks.MockIOpenRouterClient) {
				ai.EXPECT().GenAI(gomock.Any()).Return("Generated text", nil)
			},
			wantResult: "Generated text",
		},
		{
			name: "código AI ok",
			req:  request.AI{CodeAI: "HTTP server in Go"},
			setupMocks: func(ai *mocks.MockIOpenRouterClient) {
				ai.EXPECT().GenAI(gomock.Any()).Return("package main\n...", nil)
			},
			wantResult: "package main",
		},
		{
			name: "texto AI falla",
			req:  request.AI{TextAI: "something"},
			setupMocks: func(ai *mocks.MockIOpenRouterClient) {
				ai.EXPECT().GenAI(gomock.Any()).Return("", errors.New("api error"))
			},
			wantErr: true,
		},
		{
			name:       "sin input",
			req:        request.AI{},
			setupMocks: func(ai *mocks.MockIOpenRouterClient) {},
			wantResult: "No se puede autocompletar AI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, _, _, _, ai := newTestService(ctrl)
			tt.setupMocks(ai)

			result, err := svc.AutoCompleteAI(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, result, tt.wantResult)
			}
		})
	}
}

// --- SaveImage ---

func TestSaveImage(t *testing.T) {
	tests := []struct {
		name       string
		makeFile   func() (multipart.File, *multipart.FileHeader)
		setupMocks func(*mocks.MockIR2Client)
		wantErr    bool
	}{
		{
			name:     "JPEG ok",
			makeFile: makeJPEGFile,
			setupMocks: func(r2 *mocks.MockIR2Client) {
				r2.EXPECT().PutObject(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			name:     "PNG ok",
			makeFile: makePNGFile,
			setupMocks: func(r2 *mocks.MockIR2Client) {
				r2.EXPECT().PutObject(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			name:       "archivo demasiado grande",
			makeFile:   makeLargeFile,
			setupMocks: func(r2 *mocks.MockIR2Client) {},
			wantErr:    true,
		},
		{
			name:       "formato inválido",
			makeFile:   makeInvalidFile,
			setupMocks: func(r2 *mocks.MockIR2Client) {},
			wantErr:    true,
		},
		{
			name:     "R2 falla",
			makeFile: makeJPEGFile,
			setupMocks: func(r2 *mocks.MockIR2Client) {
				r2.EXPECT().PutObject(gomock.Any(), gomock.Any()).Return(nil, errors.New("R2 error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, _, _, r2, _ := newTestService(ctrl)
			tt.setupMocks(r2)

			f, header := tt.makeFile()
			_, err := svc.SaveImage(f, header)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- imageToReader ---

func TestImageToReader(t *testing.T) {
	img := makeTestImage()
	assert.NotNil(t, imageToReader("jpeg", img))
	assert.NotNil(t, imageToReader("jpg", img))
	assert.NotNil(t, imageToReader("png", img))
	assert.Nil(t, imageToReader("gif", img))
	assert.Nil(t, imageToReader("webp", img))
}

// --- helpers ---

// multipartFileReader implements multipart.File over a bytes.Reader.
type multipartFileReader struct {
	*bytes.Reader
}

func (m *multipartFileReader) Close() error { return nil }

func makeTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	return img
}

func encodeJPEG(img image.Image) *bytes.Buffer {
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, nil)
	return &buf
}

func encodePNG(img image.Image) *bytes.Buffer {
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return &buf
}

func makeJPEGFile() (multipart.File, *multipart.FileHeader) {
	buf := encodeJPEG(makeTestImage())
	return &multipartFileReader{bytes.NewReader(buf.Bytes())},
		&multipart.FileHeader{Filename: "test.jpg", Size: int64(buf.Len())}
}

func makePNGFile() (multipart.File, *multipart.FileHeader) {
	buf := encodePNG(makeTestImage())
	return &multipartFileReader{bytes.NewReader(buf.Bytes())},
		&multipart.FileHeader{Filename: "test.png", Size: int64(buf.Len())}
}

func makeLargeFile() (multipart.File, *multipart.FileHeader) {
	return &multipartFileReader{bytes.NewReader([]byte("x"))},
		&multipart.FileHeader{Filename: "big.jpg", Size: 11 << 20} // 11 MB
}

func makeInvalidFile() (multipart.File, *multipart.FileHeader) {
	data := []byte("not an image")
	return &multipartFileReader{bytes.NewReader(data)},
		&multipart.FileHeader{Filename: "bad.txt", Size: int64(len(data))}
}
