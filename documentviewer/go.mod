module p9e.in/ugcl/documentviewer

go 1.23

require (
	github.com/google/uuid v1.6.0
	github.com/klauspost/compress v1.17.0
	github.com/disintegration/imaging v1.6.2
	github.com/otiai10/gosseract/v2 v2.4.1
	github.com/gen2brain/go-fitz v1.23.7
	github.com/minio/minio-go/v7 v7.0.63
	go.uber.org/fx v1.20.0
	go.uber.org/zap v1.26.0
	google.golang.org/protobuf v1.31.0
	gorm.io/datatypes v1.2.0
	p9e.in/ugcl/packages/events v0.0.0
	p9e.in/ugcl/packages/config v0.0.0
	p9e.in/ugcl/packages/database v0.0.0
	p9e.in/ugcl/packages/vfs v0.0.0
	p9e.in/ugcl/searchservice v0.0.0
)

replace p9e.in/ugcl/packages/events => ../packages/events
replace p9e.in/ugcl/packages/config => ../packages/config
replace p9e.in/ugcl/packages/database => ../packages/database
replace p9e.in/ugcl/packages/vfs => ../packages/vfs
replace p9e.in/ugcl/searchservice => ../searchservice