/*
 * Copyright (C) 2017 Red Hat, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package dede

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type videoRecorder struct {
	filename  string
	frameRate int
	width     int
	height    int
	cancel    context.CancelFunc
}

func (v *videoRecorder) start() error {
	cmd := "ffmpeg"
	args := []string{
		"-f", "x11grab",
		"-framerate", fmt.Sprintf("%d", v.frameRate),
		"-video_size", fmt.Sprintf("%dx%d", v.width, v.height),
		"-i", ":1.0+0,0",
		"-draw_mouse", "0",
		"-segment_format_options", "movflags=+faststart",
		"-crf", "0",
		"-preset", "ultrafast",
		"-qp", "0",
		"-y",
		"-an", v.filename}

	Log.Infof("start video recording: %s %v", cmd, args)

	command := exec.Command(cmd, args...)
	if err := command.Start(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	go func() {
		<-ctx.Done()
		err := command.Process.Signal(os.Interrupt)
		if err != nil {
			Log.Errorf("cannot kill video recorder process: %s %v", cmd, args)
			return
		}
		command.Wait()
	}()
	return nil
}

func (v *videoRecorder) stop() {
	v.cancel()
}

func newVideoRecorder(filename string, width int, height int, frameRate int) *videoRecorder {
	return &videoRecorder{
		filename:  filename,
		width:     width,
		height:    height,
		frameRate: frameRate,
	}
}
