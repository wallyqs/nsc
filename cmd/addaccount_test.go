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

func Test_AddAccount(t *testing.T) {
	ts := NewTestStore(t, "add_account")
	defer ts.Done(t)

	_, bar, _ := CreateAccountKey(t)
	_, badBar, _ := CreateClusterKey(t)

	tests := CmdTests{
		{CreateAddAccountCmd(), []string{"add", "account"}, nil, []string{"account name is required"}, true},
		{CreateAddAccountCmd(), []string{"add", "account", "--name", "A"}, nil, []string{"Generated account key", "added account"}, false},
		{CreateAddAccountCmd(), []string{"add", "account", "--name", "A"}, nil, []string{"the account \"A\" already exists"}, true},
		{CreateAddAccountCmd(), []string{"add", "account", "--name", "B", "--public-key", bar}, nil, nil, false},
		{CreateAddAccountCmd(), []string{"add", "account", "--name", "X", "--public-key", badBar}, nil, []string{"invalid account key"}, true},
		{CreateAddAccountCmd(), []string{"add", "account", "--name", "badexp", "--expiry", "2018-01-01"}, nil, []string{"expiry \"2018-01-01\" is in the past"}, true},
		{CreateAddAccountCmd(), []string{"add", "account", "--name", "badexp", "--expiry", "30d"}, nil, nil, false},
	}

	tests.Run(t, "root", "add")
}

func Test_AddAccountNoStore(t *testing.T) {
	// reset the store
	ForceStoreRoot(t, "")
	ngsStore = nil
	_, _, err := ExecuteCmd(CreateAddAccountCmd())
	require.Equal(t, "no stores available", err.Error())
}

func Test_AddAccountValidateOutput(t *testing.T) {
	ts := NewTestStore(t, "test")
	defer ts.Done(t)

	_, _, err := ExecuteCmd(CreateAddAccountCmd(), "--name", "A", "--start", "2018-01-01", "--expiry", "2050-01-01")
	require.NoError(t, err)
	validateAddAccountClaims(t, ts)
}

func Test_AddAccountInteractive(t *testing.T) {
	ts := NewTestStore(t, "test")
	defer ts.Done(t)

	inputs := []interface{}{"A", true, "2018-01-01", "2050-01-01"}

	cmd := CreateAddAccountCmd()
	HoistRootFlags(cmd)
	_, _, err := ExecuteInteractiveCmd(cmd, inputs)
	require.NoError(t, err)
	validateAddAccountClaims(t, ts)
}

func validateAddAccountClaims(t *testing.T, ts *TestStore) {
	kp, err := ts.KeyStore.GetAccountKey("A")
	_, err = kp.Seed()
	require.NoError(t, err, "stored key should be a seed")

	ac, err := ts.Store.ReadAccountClaim("A")
	require.NoError(t, err, "reading account claim")

	pub, err := kp.PublicKey()
	require.Equal(t, ac.Subject, pub, "public key is subject")

	okp, err := ts.KeyStore.GetOperatorKey("test")
	require.NoError(t, err)

	if okp == nil {
		okp = kp
	}

	oppub, err := okp.PublicKey()
	require.NoError(t, err, "getting public key for operator")
	require.Equal(t, ac.Issuer, oppub, "operator signed it")

	start, err := ParseExpiry("2018-01-01")
	require.NoError(t, err)
	require.Equal(t, start, ac.NotBefore)

	expire, err := ParseExpiry("2050-01-01")
	require.NoError(t, err)
	require.Equal(t, expire, ac.Expires)
}

func Test_AddAccountOperatorLessStore(t *testing.T) {
	ts := NewTestStoreWithOperator(t, "test", nil)
	defer ts.Done(t)

	_, _, err := ExecuteCmd(CreateAddAccountCmd(), "--name", "A", "--start", "2018-01-01", "--expiry", "2050-01-01")
	require.NoError(t, err)
	validateAddAccountClaims(t, ts)
}
