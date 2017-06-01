package cloud

type Provider interface {
	Cluster() string
	Region() string
	Zones() ([]string, error)
	CreateDisk(args *CreateDiskArgs) (string, error)
	DeleteDisk(diskId string) error
}
