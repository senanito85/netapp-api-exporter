package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/pepabo/go-netapp/netapp"
	"github.com/prometheus/client_golang/prometheus"
)

type NetappVolume struct {
	ProjectID                         string
	ShareID                           string
	ShareName                         string
	FilerName                         string
	Vserver                           string
	Volume                            string
	Comment                           string
	State                             int
	Size                              int
	SizeTotal                         float64
	SizeAvailable                     float64
	SizeUsed                          float64
	SizeUsedBySnapshots               float64
	SizeAvailableForSnapshots         float64
	SnapshotReserveSize               float64
	PercentageSizeUsed                float64
	PercentageCompressionSpaceSaved   float64
	PercentageDeduplicationSpaceSaved float64
	PercentageTotalSpaceSaved         float64
}

type volumeMetrics []struct {
	desc    *prometheus.Desc
	valType prometheus.ValueType
	evalFn  func(volume *NetappVolume) float64
}

var (
	volumeLabels = []string{
		"vserver",
		"volume",
		"project_id",
		"share_id",
	}

	volMetrics = volumeMetrics{
		{
			desc: prometheus.NewDesc(
				"netapp_volume_state",
				"Netapp Volume Metrics: state (1: online; 2: restricted; 3: offline; 4: quiesced)",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return float64(v.State) },
		},
		{
			desc: prometheus.NewDesc(
				"netapp_volume_total_bytes",
				"Netapp Volume Metrics: total size",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.SizeTotal },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_used_bytes",
				"Netapp Volume Metrics: used size",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.SizeUsed },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_available_bytes",
				"Netapp Volume Metrics: available size",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.SizeAvailable },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_snapshot_used_bytes",
				"Netapp Volume Metrics: size used by snapshots",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.SizeUsedBySnapshots },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_snapshot_available_bytes",
				"Netapp Volume Metrics: size available for snapshots",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.SizeAvailableForSnapshots },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_snapshot_reserved_bytes",
				"Netapp Volume Metrics: size reserved for snapshots",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.SnapshotReserveSize },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_used_percentage",
				"Netapp Volume Metrics: used percentage ",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.PercentageSizeUsed },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_saved_total_percentage",
				"Netapp Volume Metrics: percentage of space compression and deduplication saved",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.PercentageTotalSpaceSaved },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_saved_compression_percentage",
				"Netapp Volume Metrics: percentage of space compression saved",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.PercentageCompressionSpaceSaved },
		}, {
			desc: prometheus.NewDesc(
				"netapp_volume_saved_deduplication_percentage",
				"Netapp Volume Metrics: percentage of space deduplication saved",
				volumeLabels,
				nil),
			valType: prometheus.GaugeValue,
			evalFn:  func(v *NetappVolume) float64 { return v.PercentageDeduplicationSpaceSaved },
		},
	}
)

type VolumeCollector struct {
	ApiCollectorBase
	Filer   NetappFilerClient
	Volumes []*NetappVolume
}

func (v *VolumeCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range volMetrics {
		ch <- v.desc
	}
}

func (v *VolumeCollector) Collect(ch chan<- prometheus.Metric) {
	for _, v := range v.Volumes {
		labels := []string{v.Vserver, v.Volume, v.ProjectID, v.ShareID}
		for _, m := range volMetrics {
			ch <- prometheus.MustNewConstMetric(m.desc, m.valType, m.evalFn(v), labels...)
		}
	}
}

func (v *VolumeCollector) SaveData(data []interface{}) error {
	vols := make([]*NetappVolume, 0)
	for _, d := range data {
		if v, ok := d.(*NetappVolume); ok {
			vols = append(vols, v)
		} else {
			return fmt.Errorf("type of parameter should be %s", "[]*NetappVolume")
		}
	}
	v.Volumes = vols
	return nil
}

func (v *VolumeCollector) Fetch() (volumes []interface{}, err error) {
	volumeOptions := netapp.VolumeOptions{
		MaxRecords: 20,
		DesiredAttributes: &netapp.VolumeQuery{
			VolumeInfo: &netapp.VolumeInfo{
				VolumeIDAttributes: &netapp.VolumeIDAttributes{
					Name:              "x",
					OwningVserverName: "x",
					OwningVserverUUID: "x",
					Comment:           "x",
				},
				VolumeSpaceAttributes: &netapp.VolumeSpaceAttributes{
					Size:                      1,
					SizeTotal:                 "x",
					SizeAvailable:             "x",
					SizeUsed:                  "x",
					SizeUsedBySnapshots:       "x",
					SizeAvailableForSnapshots: "x",
					SnapshotReserveSize:       "x",
					PercentageSizeUsed:        "x",
				},
				VolumeSisAttributes: &netapp.VolumeSisAttributes{
					PercentageCompressionSpaceSaved:   "x",
					PercentageDeduplicationSpaceSaved: "x",
					PercentageTotalSpaceSaved:         "x",
				},
				VolumeStateAttributes: &netapp.VolumeStateAttributes{
					State: "x",
				},
			},
		},
	}

	vols, err := v.Filer.QueryVolumes(&volumeOptions)

	if err == nil {
		logger.Printf("%s: %d volumes fetched", v.Filer.Host, len(vols))
		volumes = make([]interface{}, 0)
		for _, vol := range vols {
			nv := &NetappVolume{FilerName: v.Filer.Name}
			if vol.VolumeIDAttributes != nil {
				nv.Vserver = vol.VolumeIDAttributes.OwningVserverName
				nv.Volume = vol.VolumeIDAttributes.Name
			} else {
				// Skip if ID Attributes missing
				logger.Warnf("missing `VolumeIDAttributes` in %+v", vol)
				continue
			}
			if vol.VolumeSpaceAttributes != nil {
				v := vol.VolumeSpaceAttributes
				sizeTotal, _ := strconv.ParseFloat(v.SizeTotal, 64)
				sizeAvailable, _ := strconv.ParseFloat(v.SizeAvailable, 64)
				sizeUsed, _ := strconv.ParseFloat(v.SizeUsed, 64)
				sizeUsedBySnapshots, _ := strconv.ParseFloat(v.SizeUsedBySnapshots, 64)
				sizeAvailableForSnapshots, _ := strconv.ParseFloat(v.SizeAvailableForSnapshots, 64)
				snapshotReserveSize, _ := strconv.ParseFloat(v.SnapshotReserveSize, 64)
				percentageSizeUsed, _ := strconv.ParseFloat(v.PercentageSizeUsed, 64)

				nv.Size = vol.VolumeSpaceAttributes.Size
				nv.SizeAvailable = sizeAvailable
				nv.SizeTotal = sizeTotal
				nv.SizeUsed = sizeUsed
				nv.SizeUsedBySnapshots = sizeUsedBySnapshots
				nv.SizeAvailableForSnapshots = sizeAvailableForSnapshots
				nv.SnapshotReserveSize = snapshotReserveSize
				nv.PercentageSizeUsed = percentageSizeUsed
			} else {
				logger.Warnf("%s has no VolumeSpaceAttributes", nv.Volume)
			}
			if vol.VolumeSisAttributes != nil {
				v := vol.VolumeSisAttributes
				percentageCompressionSpaceSaved, _ := strconv.ParseFloat(v.PercentageCompressionSpaceSaved, 64)
				percentageDeduplicationSpaceSaved, _ := strconv.ParseFloat(v.PercentageDeduplicationSpaceSaved, 64)
				percentageTotalSpaceSaved, _ := strconv.ParseFloat(v.PercentageTotalSpaceSaved, 64)

				nv.PercentageCompressionSpaceSaved = percentageCompressionSpaceSaved
				nv.PercentageDeduplicationSpaceSaved = percentageDeduplicationSpaceSaved
				nv.PercentageTotalSpaceSaved = percentageTotalSpaceSaved
			} else {
				logger.Warnf("%s has no VolumeSisAttributes", vol.VolumeIDAttributes.Name)
				logger.Debugf("%+v", vol.VolumeIDAttributes)
			}
			if vol.VolumeIDAttributes.Comment != "" {
				shareID, shareName, projectID, err := parseVolumeComment(vol.VolumeIDAttributes.Comment)
				if err != nil {
					logger.Warn(err)
				} else {
					nv.ShareID = shareID
					nv.ShareName = shareName
					nv.ProjectID = projectID
				}
			} else {
				//logger.Warnf("%s (%s) does not have comment",
				//	vol.VolumeIDAttributes.Name, vol.VolumeIDAttributes.OwningVserverName)
			}
			if vol.VolumeStateAttributes != nil {
				if vol.VolumeStateAttributes.State == "online" {
					nv.State = 1
				} else if vol.VolumeStateAttributes.State == "restricted" {
					nv.State = 2
				} else if vol.VolumeStateAttributes.State == "offline" {
					nv.State = 3
				} else if vol.VolumeStateAttributes.State == "quiesced" {
					nv.State = 4
				}
			}
			volumes = append(volumes, nv)
		}
	}

	return
}

func parseVolumeComment(c string) (shareID string, shareName string, projectID string, err error) {
	r := regexp.MustCompile(`(\w+): ([\w-]+)`)
	matches := r.FindAllStringSubmatch(c, 3)
	for _, m := range matches {
		switch m[1] {
		case "share_id":
			shareID = m[2]
		case "share_name":
			shareName = m[2]
		case "project":
			projectID = m[2]
		}
	}
	if shareID == "" || projectID == "" {
		err = fmt.Errorf("failed to parse share_id/project from '%s'", c)
	}
	return
}
