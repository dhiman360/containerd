/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package sbserver

import (
	"context"
	"testing"

	runtimespec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"

	criconfig "github.com/containerd/containerd/pkg/cri/config"
)

func TestUpdateOCILinuxResource(t *testing.T) {
	oomscoreadj := new(int)
	*oomscoreadj = -500
	for desc, test := range map[string]struct {
		spec      *runtimespec.Spec
		request   *runtime.UpdateContainerResourcesRequest
		expected  *runtimespec.Spec
		expectErr bool
	}{
		"should be able to update each resource": {
			spec: &runtimespec.Spec{
				Process: &runtimespec.Process{OOMScoreAdj: oomscoreadj},
				Linux: &runtimespec.Linux{
					Resources: &runtimespec.LinuxResources{
						Memory: &runtimespec.LinuxMemory{Limit: proto.Int64(12345)},
						CPU: &runtimespec.LinuxCPU{
							Shares: proto.Uint64(1111),
							Quota:  proto.Int64(2222),
							Period: proto.Uint64(3333),
							Cpus:   "0-1",
							Mems:   "2-3",
						},
						Unified: map[string]string{"memory.min": "65536", "memory.swap.max": "1024"},
					},
				},
			},
			request: &runtime.UpdateContainerResourcesRequest{
				Linux: &runtime.LinuxContainerResources{
					CpuPeriod:          6666,
					CpuQuota:           5555,
					CpuShares:          4444,
					MemoryLimitInBytes: 54321,
					OomScoreAdj:        500,
					CpusetCpus:         "4-5",
					CpusetMems:         "6-7",
					Unified:            map[string]string{"memory.min": "1507328", "memory.swap.max": "0"},
				},
			},
			expected: &runtimespec.Spec{
				Process: &runtimespec.Process{OOMScoreAdj: oomscoreadj},
				Linux: &runtimespec.Linux{
					Resources: &runtimespec.LinuxResources{
						Memory: &runtimespec.LinuxMemory{
							Limit: proto.Int64(54321),
							Swap:  proto.Int64(54321),
						},
						CPU: &runtimespec.LinuxCPU{
							Shares: proto.Uint64(4444),
							Quota:  proto.Int64(5555),
							Period: proto.Uint64(6666),
							Cpus:   "4-5",
							Mems:   "6-7",
						},
						Unified: map[string]string{"memory.min": "1507328", "memory.swap.max": "0"},
					},
				},
			},
		},
		"should skip empty fields": {
			spec: &runtimespec.Spec{
				Process: &runtimespec.Process{OOMScoreAdj: oomscoreadj},
				Linux: &runtimespec.Linux{
					Resources: &runtimespec.LinuxResources{
						Memory: &runtimespec.LinuxMemory{Limit: proto.Int64(12345)},
						CPU: &runtimespec.LinuxCPU{
							Shares: proto.Uint64(1111),
							Quota:  proto.Int64(2222),
							Period: proto.Uint64(3333),
							Cpus:   "0-1",
							Mems:   "2-3",
						},
						Unified: map[string]string{"memory.min": "65536", "memory.swap.max": "1024"},
					},
				},
			},
			request: &runtime.UpdateContainerResourcesRequest{
				Linux: &runtime.LinuxContainerResources{
					CpuQuota:           5555,
					CpuShares:          4444,
					MemoryLimitInBytes: 54321,
					OomScoreAdj:        500,
					CpusetMems:         "6-7",
				},
			},
			expected: &runtimespec.Spec{
				Process: &runtimespec.Process{OOMScoreAdj: oomscoreadj},
				Linux: &runtimespec.Linux{
					Resources: &runtimespec.LinuxResources{
						Memory: &runtimespec.LinuxMemory{
							Limit: proto.Int64(54321),
							Swap:  proto.Int64(54321),
						},
						CPU: &runtimespec.LinuxCPU{
							Shares: proto.Uint64(4444),
							Quota:  proto.Int64(5555),
							Period: proto.Uint64(3333),
							Cpus:   "0-1",
							Mems:   "6-7",
						},
						Unified: map[string]string{"memory.min": "65536", "memory.swap.max": "1024"},
					},
				},
			},
		},
		"should be able to fill empty fields": {
			spec: &runtimespec.Spec{
				Process: &runtimespec.Process{OOMScoreAdj: oomscoreadj},
				Linux: &runtimespec.Linux{
					Resources: &runtimespec.LinuxResources{
						Memory: &runtimespec.LinuxMemory{Limit: proto.Int64(12345)},
					},
				},
			},
			request: &runtime.UpdateContainerResourcesRequest{
				Linux: &runtime.LinuxContainerResources{
					CpuPeriod:          6666,
					CpuQuota:           5555,
					CpuShares:          4444,
					MemoryLimitInBytes: 54321,
					OomScoreAdj:        500,
					CpusetCpus:         "4-5",
					CpusetMems:         "6-7",
					Unified:            map[string]string{"memory.min": "65536", "memory.swap.max": "1024"},
				},
			},
			expected: &runtimespec.Spec{
				Process: &runtimespec.Process{OOMScoreAdj: oomscoreadj},
				Linux: &runtimespec.Linux{
					Resources: &runtimespec.LinuxResources{
						Memory: &runtimespec.LinuxMemory{
							Limit: proto.Int64(54321),
							Swap:  proto.Int64(54321),
						},
						CPU: &runtimespec.LinuxCPU{
							Shares: proto.Uint64(4444),
							Quota:  proto.Int64(5555),
							Period: proto.Uint64(6666),
							Cpus:   "4-5",
							Mems:   "6-7",
						},
						Unified: map[string]string{"memory.min": "65536", "memory.swap.max": "1024"},
					},
				},
			},
		},
		"should be able to patch the unified map": {
			spec: &runtimespec.Spec{
				Process: &runtimespec.Process{OOMScoreAdj: oomscoreadj},
				Linux: &runtimespec.Linux{
					Resources: &runtimespec.LinuxResources{
						Memory: &runtimespec.LinuxMemory{Limit: proto.Int64(12345)},
						CPU: &runtimespec.LinuxCPU{
							Shares: proto.Uint64(1111),
							Quota:  proto.Int64(2222),
							Period: proto.Uint64(3333),
							Cpus:   "0-1",
							Mems:   "2-3",
						},
						Unified: map[string]string{"memory.min": "65536", "memory.max": "1507328"},
					},
				},
			},
			request: &runtime.UpdateContainerResourcesRequest{
				Linux: &runtime.LinuxContainerResources{
					CpuPeriod:          6666,
					CpuQuota:           5555,
					CpuShares:          4444,
					MemoryLimitInBytes: 54321,
					OomScoreAdj:        500,
					CpusetCpus:         "4-5",
					CpusetMems:         "6-7",
					Unified:            map[string]string{"memory.min": "1507328", "memory.swap.max": "1024"},
				},
			},
			expected: &runtimespec.Spec{
				Process: &runtimespec.Process{OOMScoreAdj: oomscoreadj},
				Linux: &runtimespec.Linux{
					Resources: &runtimespec.LinuxResources{
						Memory: &runtimespec.LinuxMemory{
							Limit: proto.Int64(54321),
							Swap:  proto.Int64(54321),
						},
						CPU: &runtimespec.LinuxCPU{
							Shares: proto.Uint64(4444),
							Quota:  proto.Int64(5555),
							Period: proto.Uint64(6666),
							Cpus:   "4-5",
							Mems:   "6-7",
						},
						Unified: map[string]string{"memory.min": "1507328", "memory.max": "1507328", "memory.swap.max": "1024"},
					},
				},
			},
		},
	} {
		t.Run(desc, func(t *testing.T) {
			config := criconfig.Config{
				PluginConfig: criconfig.PluginConfig{
					TolerateMissingHugetlbController: true,
					DisableHugetlbController:         false,
				},
			}
			got, err := updateOCIResource(context.Background(), test.spec, test.request, config)
			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}
