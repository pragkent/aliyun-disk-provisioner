package volume

import (
	"errors"
	"fmt"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/pragkent/aliyun-disk-provisioner/cloud"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kubernetes/pkg/volume"
)

const (
	annProvisionerIdentityKey = "aliyundisk-provisioner.identity"
	annCreatedByKey           = "kubernetes.io/createdby"
	annCreatedBy              = "aliyun-disk-dynamic-provisioner"
)

const (
	optionDiskId = "diskId"
)

type aliyunDiskProvisioner struct {
	identity string
	cluster  string
	client   kubernetes.Interface
	provider cloud.Provider
}

func NewProvisioner(id string, cluster string, client kubernetes.Interface, provider cloud.Provider) controller.Provisioner {
	return &aliyunDiskProvisioner{
		identity: id,
		cluster:  cluster,
		client:   client,
		provider: provider,
	}
}

func (p *aliyunDiskProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	if options.PVC.Spec.Selector != nil {
		return nil, fmt.Errorf("claim.Spec.Selector is not supported for dynamic provisioning")
	}

	diskId, sizeGB, labels, err := p.createVolume(options)
	if err != nil {
		return nil, err
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Labels:      labels,
			Annotations: p.getVolumeAnnotations(),
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse(fmt.Sprintf("%dGi", sizeGB)),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Driver:   "pragkent.me/aliyun-disk",
					FSType:   "ext4",
					Options:  p.getFlexVolumeOptions(diskId),
					ReadOnly: false,
				},
			},
		},
	}

	glog.Info("Successfully provisioned Aliyun Disk volume %s", options.PVName)
	return pv, nil
}

func (p *aliyunDiskProvisioner) createVolume(options controller.VolumeOptions) (string, int, map[string]string, error) {
	cfg, err := parseDiskConfig(options, p.provider)
	if err != nil {
		return "", 0, nil, err
	}

	zone := cfg.ChooseZoneForVolume(options.PVC.Name)
	diskName := volume.GenerateVolumeName(p.cluster, options.PVName, 128)
	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	requestBytes := capacity.Value()
	requestGB := int(volume.RoundUpSize(requestBytes, 1024*1024*1024))

	args := &cloud.CreateDiskArgs{
		Zone:         zone,
		DiskName:     diskName,
		DiskCategory: cloud.DiskCategory(cfg.Category),
		Size:         requestGB,
	}

	diskId, err := p.provider.CreateDisk(args)
	if err != nil {
		return "", 0, nil, fmt.Errorf("provider.CreateDisk error: %v", err)
	}

	return diskId, 0, p.getVolumeLabels(zone), nil
}

func (p *aliyunDiskProvisioner) getVolumeAnnotations() map[string]string {
	annotations := make(map[string]string)
	annotations[annCreatedByKey] = annCreatedBy
	return annotations
}

func (p *aliyunDiskProvisioner) getVolumeLabels(zone string) map[string]string {
	labels := make(map[string]string)
	labels[metav1.LabelZoneFailureDomain] = zone
	labels[metav1.LabelZoneRegion] = p.provider.Region()
	return labels
}

func (p *aliyunDiskProvisioner) getFlexVolumeOptions(diskId string) map[string]string {
	options := make(map[string]string)
	options[optionDiskId] = diskId
	return options
}

func (p *aliyunDiskProvisioner) Delete(volume *v1.PersistentVolume) error {
	provisionerId, ok := volume.Annotations[annProvisionerIdentityKey]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}

	if provisionerId != p.identity {
		return &controller.IgnoredError{"identity annotation on PV does not match ours"}
	}

	if err := p.deleteVolume(volume); err != nil {
		return err
	}

	glog.Info("Successfully deleted Aliyun Disk volume %s", volume.Name)
	return nil
}

func (p *aliyunDiskProvisioner) deleteVolume(volume *v1.PersistentVolume) error {
	options := volume.Spec.PersistentVolumeSource.FlexVolume.Options
	diskId, ok := options[optionDiskId]
	if !ok {
		return errors.New("diskId option not found on volume")
	}

	if err := p.provider.DeleteDisk(diskId); err != nil {
		glog.Errorf("Provider.deleteVolume error. %v. DiskId: %s", err, diskId)
		return errors.New("provider delete volume error")
	}

	return nil
}
