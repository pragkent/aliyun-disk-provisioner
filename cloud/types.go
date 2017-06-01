package cloud

import "github.com/denverdino/aliyungo/ecs"

type DiskCategory string

const (
	DiskCategoryCloud           = DiskCategory("cloud")
	DiskCategoryCloudEfficiency = DiskCategory("cloud_efficiency")
	DiskCategoryCloudSSD        = DiskCategory("cloud_ssd")
)

type CreateDiskArgs struct {
	Zone         string
	DiskName     string
	DiskCategory DiskCategory
	Size         int
	Description  string
}

type Disk ecs.DiskType
