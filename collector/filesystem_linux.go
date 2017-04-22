// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !nofilesystem

package collector

import (
	"github.com/prometheus/common/log"
	"os/exec"
	"strconv"
	"strings"
)

const (
	defIgnoredMountPoints = "^/(sys|proc|dev)($|/)"
	defIgnoredFSTypes     = "^(sys|proc|auto)fs$"
	readOnly              = 0x1 // ST_RDONLY
)

// GetStats returns filesystem stats.
func (c *filesystemCollector) GetStats() ([]filesystemStats, error) {
	log.Infof("Applatix Volume Collector!")

	mps, err := mountPointDetails()
	if err != nil {
		return nil, err
	}
	stats := []filesystemStats{}
	for _, labels := range mps {
		if c.ignoredMountPointsPattern.MatchString(labels.mountPoint) {
			log.Debugf("Ignoring mount point: %s", labels.mountPoint)
			continue
		}
		if c.ignoredFSTypesPattern.MatchString(labels.fsType) {
			log.Debugf("Ignoring fs type: %s", labels.fsType)
			continue
		}

		cmd := "nsenter"
		args := []string{"-t", "1", "-m", "df", "-T", "-B", "1", labels.mountPoint}
		cmdOut, err := exec.Command(cmd, args...).Output()

		if err != nil {
			stats = append(stats, filesystemStats{
				labels:      labels,
				deviceError: 1,
			})
			log.Errorf("Error on nsenter df call for %q: %s", labels.mountPoint, err)
			continue
		}

		outputs := strings.Split(string(cmdOut), "\n")

		if len(outputs) > 1 {
			results := strings.Fields(outputs[1])
			if len(results) > 4 {
				size_res, err1 := strconv.ParseFloat(string(results[2]), 64)
				free_res, err2 := strconv.ParseFloat(string(results[4]), 64)
				avail_res, err3 := strconv.ParseFloat(string(results[3]), 64)

				if err1 != nil || err2 != nil || err3 != nil {
					stats = append(stats, filesystemStats{
						labels:      labels,
						deviceError: 1,
					})
					log.Errorf("Error on parseFloat.")
					continue
				}
				stats = append(stats, filesystemStats{
					labels:    labels,
					size:      float64(size_res),
					free:      float64(free_res),
					avail:     float64(avail_res),
					files:     float64(0),
					filesFree: float64(0),
					ro:        0,
				})
				log.Infof(outputs[1])
			}
		}
	}
	return stats, nil
}

func mountPointDetails() ([]filesystemLabels, error) {
	filesystems := []filesystemLabels{}
	cmd := "nsenter"
	args := []string{"-t", "1", "-m", "cat", "/proc/mounts"}
	cmdOut, err := exec.Command(cmd, args...).Output()

	if err != nil {
		return nil, err
	}

	outputs := strings.Split(string(cmdOut), "\n")
	for _, output := range outputs {
		results := strings.Fields(output)
		if len(results) > 3 {
			filesystem := filesystemLabels{
				device:     string(results[0]),
				mountPoint: string(results[1]),
				fsType:     string(results[2]),
			}
			filesystems = append(filesystems, filesystem)
		}
	}
	return filesystems, nil
}
