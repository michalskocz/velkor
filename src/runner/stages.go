/*
 Copyright (c) 2026 Michał Skoczylas

 Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer
    in the documentation and/or other materials provided with the distribution.

 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
 INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
 IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
 OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA,
 OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

package runner

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/michalskocz/velkor/src/internal"

	"github.com/michalskocz/velkor/src/configuration"
)

func fetchStage() (configuration.Stage, error) {
	rawStage, err := cfg.Stages.Get()
	if err != nil {
		return configuration.Stage{}, fmt.Errorf("error getting stage: %w", err)
	}

	stage, ok := rawStage.(configuration.Stage)
	if !ok {
		return configuration.Stage{}, fmt.Errorf("invalid Stage object type")
	}

	return stage, nil
}

func runStage(stage configuration.Stage) error {
	if stage.Tasks == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workers := resolveWorkers()

	var wg sync.WaitGroup
	sem := make(chan struct{}, workers)

	errCh := make(chan error, 1)

	if err := os.MkdirAll(internal.LOG_DIR, 0755); err != nil {
		return fmt.Errorf("cannot create %s dir", internal.LOG_DIR)
	}

	for stage.Tasks.Next() {
		if ctx.Err() != nil {
			break
		}

		task, err := fetchTask(stage)
		if err != nil {
			cancel()
			return err
		}

		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}

		wg.Add(1)

		go func(t configuration.Task) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := runTask(ctx, t); err != nil {
				select {
				case errCh <- err:
					cancel()
				default:
				}
			}
		}(task)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}
