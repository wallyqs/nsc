/*
 * Copyright 2018 The NATS Authors
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_EditAccount(t *testing.T) {
	ts := NewTestStore(t, "edit account")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	ts.AddAccount(t, "B")

	tests := CmdTests{
		{createEditAccount(), []string{"edit", "account"}, nil, []string{"specify an edit option"}, true},
		{createEditAccount(), []string{"edit", "account", "--tag", "A"}, nil, []string{"an account is required"}, true},
		{createEditAccount(), []string{"edit", "account", "--tag", "A", "--account", "A"}, nil, []string{"edited account \"A\""}, false},
	}

	tests.Run(t, "root", "edit")
}

func Test_EditAccount_Tag(t *testing.T) {
	ts := NewTestStore(t, "edit account")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	_, _, err := ExecuteCmd(createEditAccount(), "--tag", "A,B,C")
	require.NoError(t, err)

	ac, err := ts.Store.ReadAccountClaim("A")
	require.NoError(t, err)
	require.NotNil(t, ac)

	require.Len(t, ac.Tags, 3)
	require.ElementsMatch(t, ac.Tags, []string{"a", "b", "c"})
}

func Test_EditAccount_RmTag(t *testing.T) {
	ts := NewTestStore(t, "edit account")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	_, _, err := ExecuteCmd(createEditAccount(), "--tag", "A,B,C")
	require.NoError(t, err)

	_, _, err = ExecuteCmd(createEditAccount(), "--rm-tag", "A,B")
	require.NoError(t, err)

	ac, err := ts.Store.ReadAccountClaim("A")
	require.NoError(t, err)
	require.NotNil(t, ac)

	require.Len(t, ac.Tags, 1)
	require.ElementsMatch(t, ac.Tags, []string{"c"})
}

func Test_EditAccount_Times(t *testing.T) {
	ts := NewTestStore(t, "edit account")
	defer ts.Done(t)

	ts.AddAccount(t, "A")

	_, _, err := ExecuteCmd(createEditAccount(), "--start", "2018-01-01", "--expiry", "2050-01-01")
	require.NoError(t, err)

	start, err := ParseExpiry("2018-01-01")
	require.NoError(t, err)

	expiry, err := ParseExpiry("2050-01-01")
	require.NoError(t, err)

	ac, err := ts.Store.ReadAccountClaim("A")
	require.NoError(t, err)
	require.NotNil(t, ac)
	require.Equal(t, start, ac.NotBefore)
	require.Equal(t, expiry, ac.Expires)
}

func Test_EditAccountMaxConns(t *testing.T) {
	ts := NewTestStore(t, "edit account")
	defer ts.Done(t)

	ts.AddAccount(t, "A")

	_, _, err := ExecuteCmd(createEditAccount(), "--conns", "10")
	require.NoError(t, err)

	ac, err := ts.Store.ReadAccountClaim("A")
	require.NoError(t, err)
	require.NotNil(t, ac)
	require.Equal(t, int64(10), ac.Limits.Conn)
}
