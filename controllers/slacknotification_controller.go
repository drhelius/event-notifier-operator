/*


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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	eventnotifierv1 "github.com/drhelius/event-notifier-operator/api/v1"
	"github.com/drhelius/event-notifier-operator/controllers/slack"
)

// SlackNotificationReconciler reconciles a SlackNotification object
type SlackNotificationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=eventnotifier.drhelius.io,resources=slacknotifications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventnotifier.drhelius.io,resources=slacknotifications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch

func (r *SlackNotificationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log

	log.Info("SlackNotification Reconciling", "Namespace", req.NamespacedName.Namespace, "Name", req.NamespacedName.Name)

	cr := &eventnotifierv1.SlackNotification{}
	err := r.Client.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	slack.Manage(cr, r.Log)

	for _, n := range slack.Notifications {
		log.Info("SlackNotification list", "Item", n)
	}

	return ctrl.Result{}, nil
}

func (r *SlackNotificationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventnotifierv1.SlackNotification{}).
		Complete(r)
}
