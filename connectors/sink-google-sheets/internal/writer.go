// Copyright 2022 Linkall Inc.
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

package internal

import (
	"context"
	"sync"
	"time"

	"github.com/vanus-labs/cdk-go/log"
)

type BufferWriter struct {
	service       *GoogleSheetService
	flushInterval time.Duration
	flushSize     int
	// sheetName: rows
	buffer map[string][][]interface{}
	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func newBufferWriter(service *GoogleSheetService, flushInterval time.Duration, flushSize int) *BufferWriter {
	return &BufferWriter{
		buffer:        map[string][][]interface{}{},
		service:       service,
		flushInterval: flushInterval,
		flushSize:     flushSize,
	}
}

func (w *BufferWriter) AppendData(sheetName string, value []interface{}) error {
	w.lock.Lock()
	w.lock.Unlock()
	values := append(w.buffer[sheetName], value)
	w.buffer[sheetName] = values
	if len(values) >= w.flushSize {
		return w.flushBySheet(sheetName)
	}
	return nil
}

func (w *BufferWriter) FlushSheet(sheetName string) error {
	w.lock.Lock()
	w.lock.Unlock()
	return w.flushBySheet(sheetName)
}

func (w *BufferWriter) flushAll() {
	w.lock.Lock()
	w.lock.Unlock()
	for sheetName := range w.buffer {
		err := w.flushBySheet(sheetName)
		if err != nil {
			log.Error("append sheet data error", map[string]interface{}{
				log.KeyError: err,
				"sheetName":  sheetName,
			})
		}
	}
}

func (w *BufferWriter) Start(ctx context.Context) {
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.flushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-w.ctx.Done():
				w.flushAll()
				return
			case <-ticker.C:
				w.flushAll()
			}
		}
	}()
}

func (w *BufferWriter) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *BufferWriter) flushBySheet(sheetName string) error {
	values, exist := w.buffer[sheetName]
	if !exist || len(values) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := w.service.appendData(ctx, sheetName, values)
	if err != nil {
		return err
	}
	delete(w.buffer, sheetName)
	return nil
}
