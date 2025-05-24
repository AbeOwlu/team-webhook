package controller

// creating a controller that watches for new nodes and adds a team label
// create an obj of type node
// create a map cache of all nodes retrieved from the API server as asw(actual state of world)
// get changes and annotate any new nodes updating the dsw(desrired state of cluster world)

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AbeOwlu/team-webhook/internal/handlers"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
)

type Node struct {
	Typemeta TypeMeta          `json:"typemeta"`
	Metadata handlers.Metadata `json:"metadata"`
}
type TypeMeta struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}
type Nodespec struct {
	ProviderID string `json:"providerID"`
}

func NodeController() error {
	// default client conn context - init for outgoing requests
	ctx, cancel := context.WithTimeout(context.Background(), 16*time.Second)
	defer cancel()

	// create inCluster auth using SA tokens rather than ~/.kube/config certs
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// create APIServer clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Remove test auth code later
	// kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	// 	panic(err)
	// }
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err)
	// }

	// current state of world node list
	// label existing nodes
	nodeList, err := clientset.CoreV1().Nodes().List(ctx, v1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		panic(err)
	}

	noLabel := true

	for _, node := range nodeList.Items {
		for label := range node.Labels {
			if label == "team" {
				fmt.Printf("Skipping resource: %v Has label %v\n", node.Name, label)
				noLabel = false
				break
			} else {
				noLabel = true
				continue
			}
		}
		if noLabel {
			// label the node
			node.Labels["team"] = "team-webhook"
			node.Labels["kubernetes.io/role"] = "webhook-non-critical"
			_, err := clientset.CoreV1().Nodes().Update(ctx, &node, v1.UpdateOptions{FieldManager: "label-webhook"})
			if err != nil {
				panic(err)
			}
			fmt.Println("Updated")
		}
	}

	time.Sleep(5 * time.Second)
	resourceVersion := nodeList.ListMeta.ResourceVersion

	watchChan := make(<-chan watch.Event)
	// watch for changes to state of world
	// go watchNodes(&resourceVersion)
	go watchNodes(ctx, &resourceVersion, clientset, watchChan)

	// listen to watch channel and label new nodes
	for {
		events := <-watchChan
		node := events.Object.(*corev1.Node)
		for label := range node.Labels {
			if label == "team" {
				fmt.Printf("Skipping resource: %v Has label %v\n", node.Name, label)
				noLabel = false
				break
			} else {
				noLabel = true
				continue
			}
		}
		if noLabel {
			node.Labels["team"] = "team-webhook"
			node.Labels["kubernetes.io/role"] = "webhook-non-critical"
			_, err := clientset.CoreV1().Nodes().Update(ctx, node, v1.UpdateOptions{FieldManager: "label-webhook"})
			if err != nil {
				panic(err)
				// can return error to caller
			}
			fmt.Println("Updated")
		}
	}
}

func watchNodes(ctx context.Context, resourceVersion *string, api *kubernetes.Clientset, watchChan <-chan (watch.Event)) {
	// sendSentinel := true
	nodeUpdated, err := api.CoreV1().Nodes().Watch(ctx, v1.ListOptions{
		// SendInitialEvents:    &sendSentinel,
		// ResourceVersionMatch: v1.ResourceVersionMatchNotOlderThan,
		// ResourceVersion:      *resourceVersion
	})
	if err != nil {
		if strings.Contains(err.Error(), "Gone") {
			fmt.Println("Resource version is too old: %v", err.Error())
			newNodeWatch, err := api.CoreV1().Nodes().List(ctx, v1.ListOptions{})
			if err != nil {
				panic(err)
			} else {
				watchNodes(ctx, &newNodeWatch.ResourceVersion, api, watchChan)
			}
		}
	} else {
		watchChan = nodeUpdated.ResultChan()
	}
}
