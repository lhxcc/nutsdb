// Copyright 2023 The nutsdb Author. All rights reserved.
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

package nutsdb

import (
	"time"

	"github.com/antlabs/timer"
)

type nodesInBucket map[string]timer.TimeNoder // key to timer node

func newNodesInBucket() nodesInBucket {
	return make(map[string]timer.TimeNoder)
}

type nodes map[string]nodesInBucket // bucket to nodes that in a bucket

func (n nodes) getNode(bucket, key string) (timer.TimeNoder, bool) {
	nib, ok := n[bucket]
	if !ok {
		return nil, false
	}
	node, ok := nib[key]
	return node, ok
}

func (n nodes) addNode(bucket, key string, node timer.TimeNoder) {
	nib, ok := n[bucket]
	if !ok {
		nib = newNodesInBucket()
		n[bucket] = nib
	}
	nib[key] = node
}

func (n nodes) delNode(bucket, key string) {
	if nib, ok := n[bucket]; ok {
		delete(nib, key)
	}
}

type ttlManager struct {
	t          timer.Timer
	timerNodes nodes
}

func newTTLManager(expiredDeleteType ExpiredDeleteType) *ttlManager {
	var t timer.Timer

	switch expiredDeleteType {
	case TimeWheel:
		t = timer.NewTimer(timer.WithTimeWheel())
	case TimeHeap:
		t = timer.NewTimer(timer.WithMinHeap())
	default:
		t = timer.NewTimer()
	}

	return &ttlManager{
		t:          t,
		timerNodes: make(nodes),
	}
}

func (tm *ttlManager) run() {
	tm.t.Run()
}

func (tm *ttlManager) exist(bucket, key string) bool {
	_, ok := tm.timerNodes.getNode(bucket, key)
	return ok
}

func (tm *ttlManager) add(bucket, key string, expire time.Duration, callback func()) {
	if node, ok := tm.timerNodes.getNode(bucket, key); ok {
		node.Stop()
	}

	node := tm.t.AfterFunc(expire, callback)
	tm.timerNodes.addNode(bucket, key, node)
}

func (tm *ttlManager) del(bucket, key string) {
	tm.timerNodes.delNode(bucket, key)
}

func (tm *ttlManager) close() {
	tm.timerNodes = nil
	tm.t.Stop()
}
