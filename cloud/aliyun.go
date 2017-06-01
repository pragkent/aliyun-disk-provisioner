package cloud

import (
	"net/http"
	"time"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/metadata"
	"github.com/golang/glog"
)

type aliyunProvider struct {
	region    common.Region
	vpcId     string
	ecsClient *ecs.Client
	vpcClient *ecs.Client
}

func NewProvider(accessKey, secretKey, region string) Provider {
	cRegion := common.Region(region)

	return &aliyunProvider{
		region:    cRegion,
		ecsClient: ecs.NewECSClient(accessKey, secretKey, cRegion),
		vpcClient: ecs.NewVPCClient(accessKey, secretKey, cRegion),
	}
}

func (p *aliyunProvider) Cluster() string {
	vpcId := p.probeVpcId()
	if vpcId == "" {
		glog.Errorf("Instance not in VPC")
		return ""
	}

	args := &ecs.DescribeVpcsArgs{
		VpcId:    vpcId,
		RegionId: p.region,
	}

	vpcs, _, err := p.vpcClient.DescribeVpcs(args)
	if err != nil {
		glog.Errorf("DescribeVpcs error. %v. VpcId: %s", err, vpcId)
		return ""
	}

	if len(vpcs) == 0 {
		glog.Errorf("VPC does not exist. VpcId: %s", vpcId)
		return ""
	}

	return vpcs[0].VpcName
}

func (p *aliyunProvider) probeVpcId() string {
	md := metadata.NewMetaData(&http.Client{
		Timeout: 1 * time.Second,
	})

	vpcId, err := md.VpcID()
	if err != nil {
		glog.Errorf("Metadata.VpcID error. %v.", err)
		return ""
	}

	glog.Infof("Metadata.VpcID passed. VpcId: %s", vpcId)
	return vpcId
}

func (p *aliyunProvider) Region() string {
	return string(p.region)
}

func (p *aliyunProvider) Zones() ([]string, error) {
	var zones []string

	zs, err := p.ecsClient.DescribeZones(p.region)
	if err != nil {
		glog.Errorf("DescribeZones error: %v", err)
		return nil, err
	}

	for _, z := range zs {
		zones = append(zones, z.ZoneId)
	}

	return zones, nil
}

func (p *aliyunProvider) CreateDisk(args *CreateDiskArgs) (string, error) {
	ecsArgs := &ecs.CreateDiskArgs{
		ZoneId:       args.Zone,
		DiskName:     args.DiskName,
		DiskCategory: ecs.DiskCategory(args.DiskCategory),
		Size:         args.Size,
		Description:  args.Description,
	}

	diskId, err := p.ecsClient.CreateDisk(ecsArgs)
	if err != nil {
		return "", err
	}

	if err := p.ecsClient.WaitForDisk(p.region, diskId, ecs.DiskStatusAvailable, 0); err != nil {
		glog.Errorf("WaitForDisk error. %v. DiskId: %s", err, diskId)
		return "", err
	}

	return diskId, nil
}

func (p *aliyunProvider) DeleteDisk(diskId string) error {
	if err := p.ecsClient.DeleteDisk(diskId); err != nil {
		return err
	}

	return nil
}
