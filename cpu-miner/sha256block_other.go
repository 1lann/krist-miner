// +build !386,!amd64,!arm64

// Minio Cloud Storage, (C) 2016 Minio, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

func mineAVX(proc int) {
	panic("optimization unsupported")
}

func mineAVX2(proc int) {
	panic("optimization unsupported")
}

func mineSHA(proc int) {
	panic("optimization unsupported")
}

func mineSSSE3(proc int) {
	panic("optimization unsupported")
}

func mineARM(proc int) {
	panic("optimization unsupported")
}
