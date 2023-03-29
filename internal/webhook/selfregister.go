/*
 * Copyright (c) 2022 Cisco and/or its affiliates.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhook

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Register registers MutatingWebhookConfiguration
func (w *MsmWebhook) Register(ctx context.Context) error {
	w.Log.Infof("Registering MutatingWebhookConfiguration")
	defer w.Log.Infof("Successfully registered MutatingWebhookConfiguration")

	path := "/mutate"
	policy := admissionv1.Fail
	sideEffects := admissionv1.SideEffectClassNone

	webhookConfig := &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: msmName,
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name: fmt.Sprintf("%v.%v", msmName, msmAnnotation),
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{admissionv1.Create, admissionv1.Update},
						Rule: admissionv1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
					{
						Operations: []admissionv1.OperationType{admissionv1.Create, admissionv1.Update},
						Rule: admissionv1.Rule{
							APIGroups:   []string{"extensions"},
							APIVersions: []string{"v1"},
							Resources:   []string{"deployments"},
						},
					},
				},
				SideEffects:             &sideEffects,
				AdmissionReviewVersions: []string{"v1"},
				FailurePolicy:           &policy,
				ClientConfig: admissionv1.WebhookClientConfig{
					Service: &admissionv1.ServiceReference{
						Namespace: w.namespace,
						Name:      msmServiceName,
						Path:      &path,
					},
					CABundle: w.caBundle,
				},
			},
		},
	}
	_, err := w.client.MutatingWebhookConfigurations().Create(ctx, webhookConfig, metav1.CreateOptions{})

	return err
}

// Unregister unregisters MutatingWebhookConfiguration
func (w *MsmWebhook) Unregister(ctx context.Context) error {
	w.Log.Infof("Unregistering MutatingWebhookConfiguration")
	defer w.Log.Infof("Successfully unregistered MutatingWebhookConfiguration")

	return w.client.MutatingWebhookConfigurations().Delete(ctx, msmName, metav1.DeleteOptions{})
}

func (w *MsmWebhook) selfSignedCert() tls.Certificate {
	now := time.Now()

	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("mediastreamingmesh.%v-ca", msmServiceName),
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(1, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		DNSNames: []string{
			fmt.Sprintf("%v.%v", msmServiceName, w.namespace),
			fmt.Sprintf("%v.%v.svc", msmServiceName, w.namespace),
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		panic(err.Error())
	}

	certRaw, err := x509.CreateCertificate(rand.Reader, template, template, privateKey.Public(), privateKey)

	if err != nil {
		panic(err.Error())
	}

	pemCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})

	pemKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	result, err := tls.X509KeyPair(pemCert, pemKey)

	if err != nil {
		panic(err.Error())
	}

	w.caBundle = pemCert
	return result
}
