/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package sampwdapi sampwd api
package sampwdapi

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"io"

	"hcm/pkg/kit"
	"hcm/pkg/rest"
	apigateway "hcm/pkg/thirdparty/api-gateway"
)

// UpdateHostPwd 更新密码库中的主机密码
// @doc https://bkapigw.woa.com/docs/apigw-api/bk-sam-pwd/update_database_password/doc?stage=prod
func (s *samPwdCli) UpdateHostPwd(kt *kit.Kit, req *UpdateHostPwdReq) (*UpdateHostPwdResp, error) {
	encryPwd, err := encrypt(req.Password, s.config.BkToken)
	if err != nil {
		return nil, err
	}
	req.Password = encryPwd

	return apigateway.ApiGatewayCall[UpdateHostPwdReq, UpdateHostPwdResp](s.client, s.config, rest.POST,
		kt, req, "/pwd/update_database_password")
}

// encrypt plaintext with the passphrase
func encrypt(plaintext string, passphrase string) (string, error) {
	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	key, iv := deriveKeyAndIv(passphrase, string(salt))

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	pad := pkcs7Padding([]byte(plaintext), block.BlockSize())
	ecb := cipher.NewCBCEncrypter(block, []byte(iv))
	encrypted := make([]byte, len(pad))
	ecb.CryptBlocks(encrypted, pad)

	return base64.StdEncoding.EncodeToString([]byte("Salted__" + string(salt) + string(encrypted))), nil
}

// pkcs7Padding pkcs padding
func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// deriveKeyAndIv derive key and iv
func deriveKeyAndIv(passphrase string, salt string) (string, string) {
	salted := ""
	dI := ""

	for len(salted) < 48 {
		md := md5.New()
		md.Write([]byte(dI + passphrase + salt))
		dM := md.Sum(nil)
		dI = string(dM[:16])
		salted = salted + dI
	}

	key := salted[0:32]
	iv := salted[32:48]

	return key, iv
}
