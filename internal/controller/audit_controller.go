/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	reportv1 "github.com/btwseeu78/namespace-reporter/api/v1"
	"github.com/go-logr/logr"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// AuditReconciler reconciles a Audit object
type AuditReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

type HealthStatus struct {
	ReadinessProbe bool `json:"readinessProbe"`
	LivenessProbe  bool `json:"livenessProbe"`
}

type Embed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
}

type WebhookMessage struct {
	Username string   `json:"username"`
	Content  string   `json:"content"`
	Embeds   []*Embed `json:"embeds"`
}

//+kubebuilder:rbac:groups=report.arpan.io,resources=audits,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=report.arpan.io,resources=audits/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=report.arpan.io,resources=audits/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
//+kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Audit object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *AuditReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("audit", req.NamespacedName)
	log.Info("reconciling audit")

	//Get The Objects
	var audit reportv1.Audit
	if err := r.Get(ctx, req.NamespacedName, &audit); err != nil {
		log.Error(err, "unable to fetch audit")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	namespacelist := v1.NamespaceList{}

	err := r.List(ctx, &namespacelist, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(audit.Spec.Selector),
	})

	if err != nil {
		log.Error(err, "unable to list namespaces as per provided level", "name", req.Name, "labels", audit.Spec.Selector)
		return ctrl.Result{}, err
	}

	// Filter Out Conditions As per the requirements

	for _, namespace := range namespacelist.Items {
		deployList := &v12.DeploymentList{}
		helathStatus := HealthStatus{
			ReadinessProbe: false,
			LivenessProbe:  false,
		}
		err = r.List(ctx, deployList, &client.ListOptions{
			Namespace: namespace.Name,
		})
		if err != nil {
			log.Error(err, "unable to list deployments", "namespace", namespace.Name)
			continue
		}
		log.Info("found deployments", "namespace", namespace.Name, "deployments", len(deployList.Items))

		for _, deploy := range deployList.Items {
			deployment := &v12.Deployment{}
			if err := r.Get(ctx, types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, deployment); err != nil {
				log.Error(err, "Unable to list deploy", "namespace", namespace.Name, "deploy", deploy.Name)
			}
			log.Info("found deployment", "namespace", namespace.Name, "deploy", deploy.Name, "healthSpec", deployment.Spec.Template.Spec.Containers[0].Image)
			if deployment.Spec.Template.Spec.Containers[0].ReadinessProbe != nil {
				helathStatus.ReadinessProbe = true

			}
			if deployment.Spec.Template.Spec.Containers[0].LivenessProbe != nil {
				helathStatus.LivenessProbe = true
			}
			if audit.Spec.WebhookUrl != nil {
				if err := r.sendMessage(helathStatus, audit.Spec.WebhookUrl); err != nil {
					log.Error(err, "unable to send message to webhook", "namespace", namespace.Name, "deploy", deploy.Name)
					return ctrl.Result{RequeueAfter: time.Hour}, err
				}

			}
		}

	}

	return ctrl.Result{RequeueAfter: time.Hour * 24}, nil
}

func (r *AuditReconciler) sendMessage(hcheck HealthStatus, url *string) error {
	message := &WebhookMessage{
		Username: "AuditController",
		Content:  "Health Check Status",
		Embeds: []*Embed{
			{
				Title:       "Health Check Status" + time.DateTime,
				Description: fmt.Sprintf("liveness: %v,readiness: %v", hcheck.LivenessProbe, hcheck.LivenessProbe),
				Color:       16734296,
			},
		},
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = http.Post(*url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	return nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *AuditReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&reportv1.Audit{}).
		Complete(r)
}

func (r *AuditReconciler) ProcessDeployment(ctx context.Context, req ctrl.Request, nslist v1.NamespaceList) error {
	panic("error")
}
