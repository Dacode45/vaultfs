// Copyright © 2016 Asteris, LLC
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

package fs

import (
	"errors"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
)

// VaultFS is a vault filesystem
type VaultFS struct {
	*api.Client
	root       string
	conn       *fuse.Conn
	mountpoint string
}

// New returns a new VaultFS
func New(config *api.Config, mountpoint, token, root string) (*VaultFS, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	return &VaultFS{
		Client:     client,
		root:       root,
		mountpoint: mountpoint,
	}, nil
}

// Mount the FS at the given mountpoint
func (v *VaultFS) Mount() error {
	var err error
	v.conn, err = fuse.Mount(
		v.mountpoint,
		fuse.FSName("vault"),
		fuse.VolumeName("vault"),
	)

	logrus.Debug("created conn")
	if err != nil {
		return err
	}

	logrus.Debug("starting to serve")
	return fs.Serve(v.conn, v)
}

// Unmount the FS
func (v *VaultFS) Unmount() error {
	if v.conn == nil {
		return errors.New("not mounted")
	}

	err := v.conn.Close()
	if err != nil {
		return err
	}

	err = fuse.Unmount(v.mountpoint)
	if err != nil {
		return err
	}

	logrus.Debug("closed connection, waiting for ready")
	<-v.conn.Ready
	if v.conn.MountError != nil {
		return v.conn.MountError
	}

	return nil
}

// Root returns the struct that does the actual work
func (v *VaultFS) Root() (fs.Node, error) {
	logrus.Debug("returning root")
	return NewRoot(v.root, v.Logical()), nil
}
