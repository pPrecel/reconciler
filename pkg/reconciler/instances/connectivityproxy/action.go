package connectivityproxy

import (
	"strings"

	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/instances/connectivityproxy/connectivityclient"
	"github.com/kyma-incubator/reconciler/pkg/reconciler/service"
	"github.com/pkg/errors"
)

type CustomAction struct {
	Name     string
	Loader   Loader
	Commands Commands
}

func (a *CustomAction) Run(context *service.ActionContext) error {
	context.Logger.Debug("Staring invocation of " + context.Task.Component + " reconciliation")

	host := context.KubeClient.GetHost()
	if host == "" {
		return errors.Errorf("Host cannot be empty")
	}
	context.Task.Configuration["global.kubeHost"] = strings.TrimPrefix(host, "https://")

	if context.Task.Type == model.OperationTypeDelete {
		context.Logger.Debug("Requested cluster removal - removing component")
		if err := a.Commands.Remove(context); err != nil {
			context.Logger.Error("Failed to remove Connectivity Proxy: %v", err)
			return err
		}
		return nil
	}

	context.Logger.Debug("Checking StatefulSet")
	app, err := context.KubeClient.GetStatefulSet(context.Context, context.Task.Component, context.Task.Namespace)
	if err != nil {
		return errors.Wrap(err, "Error while retrieving StatefulSet")
	}

	context.Logger.Debug("Checking BTP Operator binding")
	binding, err := a.Loader.FindBindingOperator(context)
	if err != nil {
		return errors.Wrap(err, "Error while retrieving binding from BTP Operator")
	}

	if binding != nil {
		context.Logger.Debug("Reading ServiceBinding Secret")
		bindingSecret, err := a.Loader.FindSecret(context, binding)

		context.Logger.Debug("Service Binding Secret check")
		if err != nil {
			return errors.Wrap(err, "Error while retrieving service binding secret")
		}

		// build overrides for credential secret by reading them from btp-operator secret
		context.Logger.Debug("Populating configs")
		a.Commands.PopulateConfigs(context, bindingSecret)

		caClient, err := connectivityclient.NewConnectivityCAClient(context.Task.Configuration)

		if err != nil {
			return errors.Wrap(err, "Error - cannot create Connectivity CA client")
		}
		context.Logger.Debug("Creating Istio CA cacert secret for Connectivity Proxy")
		err = a.Commands.CreateCARootSecret(context, caClient)

		if err != nil {
			return errors.Wrap(err, "error during creatiion of Istio CA cacert secret for Connectivity Proxy")
		}

		refresh := app != nil

		if refresh {
			context.Logger.Info("Reconciling component")
		} else {
			context.Logger.Info("Installing component")
		}

		if err := a.Commands.Apply(context, refresh); err != nil {
			return errors.Wrap(err, "Error during reconcilation")
		}
	} else if binding == nil && app != nil {
		context.Logger.Info("Removing component")
		if err := a.Commands.Remove(context); err != nil {
			context.Logger.Error("Failed to remove Connectivity Proxy: %v", err)
			return err
		}
	}

	return nil
}
