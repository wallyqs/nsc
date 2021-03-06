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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nats-io/jwt"
	"github.com/stretchr/testify/require"
)

func Test_GenerateActivation(t *testing.T) {
	ts := NewTestStore(t, "gen activation")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	ts.AddExport(t, "A", jwt.Stream, "foo.>", false)

	_, pub, _ := CreateAccountKey(t)

	tests := CmdTests{
		{createGenerateActivationCmd(), []string{"generate", "activation"}, nil, []string{"target-account cannot be empty"}, true},
		{createGenerateActivationCmd(), []string{"generate", "activation", "--target-account", pub}, []string{"-----BEGIN NATS ACTIVATION JWT-----"}, nil, false},
	}

	tests.Run(t, "root", "generate")
}

func Test_GenerateActivationMultiple(t *testing.T) {
	ts := NewTestStore(t, "gen activation")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	ts.AddExport(t, "A", jwt.Stream, "foo.>", false)
	ts.AddExport(t, "A", jwt.Stream, "bar.>", false)
	ts.AddAccount(t, "B")

	_, pub, _ := CreateAccountKey(t)

	tests := CmdTests{
		{createGenerateActivationCmd(), []string{"generate", "activation"}, nil, []string{"an account is required"}, true},
		{createGenerateActivationCmd(), []string{"generate", "activation", "--account", "A"}, nil, []string{"a subject is required"}, true},
		{createGenerateActivationCmd(), []string{"generate", "activation", "--account", "A", "--subject", "bar.>"}, nil, []string{"target-account cannot be empty"}, true},
		{createGenerateActivationCmd(), []string{"generate", "activation", "--account", "A", "--subject", "bar.>", "--target-account", pub}, []string{"-----BEGIN NATS ACTIVATION JWT-----"}, nil, false},
	}

	tests.Run(t, "root", "generate")
}

func Test_GenerateActivationEmptyExports(t *testing.T) {
	ts := NewTestStore(t, "gen activation")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	_, _, err := ExecuteCmd(createGenerateActivationCmd())
	require.Error(t, err)
	require.Equal(t, "account \"A\" doesn't have exports", err.Error())
}

func Test_GenerateActivationNoPrivateExports(t *testing.T) {
	ts := NewTestStore(t, "gen activation")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	ts.AddExport(t, "A", jwt.Service, "foo", true)

	_, _, err := ExecuteCmd(createGenerateActivationCmd())
	require.Error(t, err)
	require.Equal(t, "account \"A\" doesn't have exports that require token generation", err.Error())
}

func Test_GenerateActivationOutputsFile(t *testing.T) {
	ts := NewTestStore(t, "gen activation")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	ts.AddExport(t, "A", jwt.Service, "foo", false)

	_, pub, _ := CreateAccountKey(t)

	outpath := filepath.Join(ts.Dir, "token.jwt")
	_, _, err := ExecuteCmd(createGenerateActivationCmd(), "--target-account", pub, "--output-file", outpath)
	require.NoError(t, err)
	testExternalToken(t, outpath)
}

func testExternalToken(t *testing.T, tokenpath string) {
	_, err := os.Stat(tokenpath)
	require.NoError(t, err)

	d, err := ioutil.ReadFile(tokenpath)
	require.NoError(t, err)

	s := ExtractToken(string(d))

	ac, err := jwt.DecodeActivationClaims(s)
	if err != nil && strings.Contains(err.Error(), "illegal base64") {
		t.Log("failed decoding a claim")
		t.Log("Extracted token\n", s)
		t.Log("Token file", tokenpath)
	}
	require.NoError(t, err)
	require.Equal(t, "foo", string(ac.ImportSubject))
}

func Test_InteractiveGenerate(t *testing.T) {
	ts := NewTestStore(t, "gen activation")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	ts.AddExport(t, "A", jwt.Service, "foo", false)

	cmd := createGenerateActivationCmd()
	HoistRootFlags(cmd)

	_, pub, _ := CreateAccountKey(t)

	outpath := filepath.Join(ts.Dir, "token.jwt")
	inputs := []interface{}{0, pub, "0", "0"}
	_, _, err := ExecuteInteractiveCmd(cmd, inputs, "-i", "--output-file", outpath)
	require.NoError(t, err)

	testExternalToken(t, outpath)
}

func Test_InteractiveExternalKeyGenerate(t *testing.T) {
	ts := NewTestStore(t, "gen activation")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	ts.AddExport(t, "A", jwt.Service, "foo", false)

	cmd := createGenerateActivationCmd()
	HoistRootFlags(cmd)

	outpath := filepath.Join(ts.Dir, "token.jwt")

	_, pub, _ := CreateAccountKey(t)
	keyfile := filepath.Join(ts.Dir, "key")
	err := ioutil.WriteFile(keyfile, []byte(pub), 0700)
	require.NoError(t, err)

	inputs := []interface{}{0, keyfile, "0", "0"}
	_, _, err = ExecuteInteractiveCmd(cmd, inputs, "-i", "--output-file", outpath)
	require.NoError(t, err)

	testExternalToken(t, outpath)
}

func Test_InteractiveMultipleAccountsGenerate(t *testing.T) {
	ts := NewTestStore(t, "gen activation")
	defer ts.Done(t)

	ts.AddAccount(t, "A")
	ts.AddExport(t, "A", jwt.Service, "foo", false)
	ts.AddAccount(t, "B")

	cmd := createGenerateActivationCmd()
	HoistRootFlags(cmd)

	outpath := filepath.Join(ts.Dir, "token.jwt")

	_, pub, _ := CreateAccountKey(t)

	inputs := []interface{}{0, 0, pub, "0", "0"}
	_, _, err := ExecuteInteractiveCmd(cmd, inputs, "-i", "--output-file", outpath)
	require.NoError(t, err)

	testExternalToken(t, outpath)
}
