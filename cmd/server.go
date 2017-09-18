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

package cmd

import (
	"github.com/skydive-project/dede/dede"
	"github.com/spf13/cobra"
)

var (
	port    int
	dataDir string
)

var server = &cobra.Command{
	Use:          "server",
	Short:        "dede server",
	Long:         "dede server",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		dede.InitServer(dataDir, port)
		dede.ListenAndServe()
	},
}

func init() {
	server.Flags().StringVarP(&dataDir, "data-dir", "", "/tmp", "data dir path, place where the files will go")
	server.Flags().IntVarP(&port, "port", "", 12345, "port used by the DeDe server, default: 12345")
}
