package webhook

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
)

// New creates a new Server with the provided options
func New(opts ...Option) *MsmWebhook {
	w := &MsmWebhook{}

	for _, o := range opts {
		o(w)
	}

	return w
}

// Option is a function that acts on a Server to inject Dependencies or configuration
type Option func(w *MsmWebhook)

// UseDeps returns Option that can inject custom dependencies.
func UseDeps(cb func(*Deps)) Option {
	return func(p *MsmWebhook) {
		cb(&p.Deps)
	}
}

// MsmWebhook holds the data structures for the webhook
type MsmWebhook struct {
	Deps

	server       *http.Server
	deserializer runtime.Decoder
}

// Deps list dependencies for the Server
type Deps struct {
	Log *logrus.Logger
}

// Init initializes the server
func (w *MsmWebhook) Init() error {
	w.Log.Info("Initializing server")

	runtimeScheme := runtime.NewScheme()
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1.AddToScheme(runtimeScheme)
	//	_ = v1.AddToScheme(runtimeScheme)

	w.deserializer = serializer.NewCodecFactory(runtimeScheme).UniversalDeserializer()

	// read and parse the certificates
	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		w.Log.Errorf("Failed to load key pair: %v", err)
		return err
	}

	w.server = &http.Server{
		Addr: fmt.Sprintf(":%v", defaultPort),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{pair},
		},
	}
	// http server and server handler initialization
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", w.handle)
	w.server.Handler = mux

	return nil
}

func (w *MsmWebhook) Start() error {
	return w.server.ListenAndServeTLS("", "")
}

func (w *MsmWebhook) Close() {
	_ = w.server.Close()
}
