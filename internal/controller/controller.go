/*
Copyright 2025 The Crossplane Authors.

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
	ctrl "sigs.k8s.io/controller-runtime"

	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"

	"github.com/rossigee/provider-rabbitmq/internal/controller/vhost"
	"github.com/rossigee/provider-rabbitmq/internal/controller/exchange"
	"github.com/rossigee/provider-rabbitmq/internal/controller/queue"
	"github.com/rossigee/provider-rabbitmq/internal/controller/binding"
	"github.com/rossigee/provider-rabbitmq/internal/controller/user"
	"github.com/rossigee/provider-rabbitmq/internal/controller/permission"
)

// Setup creates all RabbitMQ controllers and adds them to the manager.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	if err := vhost.Setup(mgr, o); err != nil {
		return err
	}
	if err := exchange.Setup(mgr, o); err != nil {
		return err
	}
	if err := queue.Setup(mgr, o); err != nil {
		return err
	}
	if err := binding.Setup(mgr, o); err != nil {
		return err
	}
	if err := user.Setup(mgr, o); err != nil {
		return err
	}
	if err := permission.Setup(mgr, o); err != nil {
		return err
	}
	return nil
}
