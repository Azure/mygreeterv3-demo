targetScope = 'subscription'

@sys.description('The name for the resources.')
param resourcesName string

@sys.description('The subscription the resources are deployed to.')
param subscriptionId string

@sys.description('The location of the resource group the resources are deployed to.')
param location string

@sys.description('The name of the resource group the resources are deployed to.')
param resourceGroupName string

// This resource is shared and defined in resources/Main.SharedResources.Template.bicep in shared-resources directory; we only reference it here. Do not remove `existing` syntax.
resource rg 'Microsoft.Resources/resourceGroups@2021-04-01' existing = {
  name: resourceGroupName
  scope: subscription(subscriptionId)
}

// This resource is shared and defined in resources/Main.SharedResources.Template.bicep in shared-resources directory; we only reference it here. Do not remove `existing` syntax.
resource serviceBusNamespace 'Microsoft.ServiceBus/namespaces@2022-10-01-preview' existing = {
  name: 'servicehubval-${resourcesName}-${location}-sb-ns'
  scope: resourceGroup(subscriptionId, resourceGroupName)
}

resource aks 'Microsoft.ContainerService/managedClusters@2024-09-02-preview' existing = {
  name: 'servicehubval-${resourcesName}-cluster'
  scope: resourceGroup(subscriptionId, resourceGroupName)
}

var serverServiceAccountNamespace = 'servicehubval-mygreeterv3-server'
var serverServiceAccountName = 'servicehubval-mygreeterv3-server'
var asyncServiceAccountNamespace = 'servicehubval-mygreeterv3-async'
var asyncServiceAccountName = 'servicehubval-mygreeterv3-async'
module managedIdentity 'br/public:avm/res/managed-identity/user-assigned-identity:0.2.1' = {
  name: 'servicehubval-${resourcesName}-mygreeterv3-managed-identityDeploy'
  scope: resourceGroup(subscriptionId, resourceGroupName)
  params: {
    // Name needs to be unique in the entire subscription, thus why we add the `${resourcesName}` to avoid conflicts from different developers.
    name: 'servicehubval-${resourcesName}-mygreeterv3-managedIdentity'
    location: rg.location
    federatedIdentityCredentials: [
      {
        name: 'servicehubval-${resourcesName}-mygreeterv3-fedIdentity-server'
        issuer: aks.properties.oidcIssuerProfile.issuerURL
        subject: 'system:serviceaccount:${serverServiceAccountNamespace}:${serverServiceAccountName}'
        audiences: ['api://AzureADTokenExchange']
      }
      {
        name: 'servicehubval-${resourcesName}-mygreeterv3-fedIdentity-async'
        issuer: aks.properties.oidcIssuerProfile.issuerURL
        subject: 'system:serviceaccount:${asyncServiceAccountNamespace}:${asyncServiceAccountName}'
        audiences: ['api://AzureADTokenExchange']
      }
    ]
  }
}

// TODO: Migrate to use bicep module registry. Current bicep registry module is management group scoped but we use subscription scoped.
module azureSdkRoleAssignment 'br:servicehubregistry.azurecr.io/bicep/modules/subscription-role-assignment:v6' = {
  name: 'servicehubval-mygreeterv3azuresdkra${location}Deploy'
  scope: subscription(subscriptionId)
  params: {
    principalId: managedIdentity.outputs.principalId
    description: 'servicehubval-mygreeterv3-${resourcesName}-contributor-azuresdk-role-assignment'
    roleDefinitionIdOrName: 'Contributor'
    principalType: 'ServicePrincipal'
    subscriptionId: subscriptionId
  }
}

module resourceRoleAssignmentServiceBusSender 'br/public:avm/ptn/authorization/resource-role-assignment:0.1.1' = {
  name: 'resourceRoleAssignmentServiceBusSenderDeployment'
  scope: resourceGroup(subscriptionId, resourceGroupName)
  params: {
    // Required parameters
    principalId: managedIdentity.outputs.principalId
    resourceId: serviceBusNamespace.id
    roleDefinitionId: '090c5cfd-751d-490a-894a-3ce6f1109419'
    // Non-required parameters
    description: 'Assign Service Bus Data Sender permissions to managed identity.'
    principalType: 'ServicePrincipal'
    roleName: 'Service Bus Data Sender'
  }
}

module resourceRoleAssignmentServiceBusReceiver 'br/public:avm/ptn/authorization/resource-role-assignment:0.1.1' = {
  name: 'resourceRoleAssignmentDeploymentServiceBusReceiver'
  scope: resourceGroup(subscriptionId, resourceGroupName)
  params: {
    // Required parameters
    principalId: managedIdentity.outputs.principalId
    resourceId: serviceBusNamespace.id
    roleDefinitionId: '4f6d3b9b-027b-4f4c-9142-0e5a2a2247e0'
    // Non-required parameters
    description: 'Assign Service Bus Data Receiver permissions to managed identity.'
    principalType: 'ServicePrincipal'
    roleName: 'Service Bus Data Receiver'
  }
}

//TODO(mheberling): SQL server can only add other users to the db (after the admin is set) via SQL users.
// Look into using SQL Managed instance or setting the admin managed identity to the pods.
module server 'br/public:avm/res/sql/server:0.9.1' = {
  name: 'mygreeterv3-${resourcesName}-serverDeploy'
  scope: resourceGroup(subscriptionId, resourceGroupName)
  params: {
    // Required parameters
    name: 'mygreeterv3-${resourcesName}-${location}-sql-server'
    location: rg.location
    // Non-required parameters
    administrators: {
      azureADOnlyAuthentication: true
      login: 'myspn'
      principalType: 'Application'
      sid: managedIdentity.outputs.clientId
    }
    databases: [
      {
        name: 'mygreeterv3-${resourcesName}-sql-database'
        zoneRedundant: false
      }
    ]
    firewallRules: [
      {
        name: 'AllowAllWindowsAzureIps'
        endIpAddress: '0.0.0.0'
        startIpAddress: '0.0.0.0'
      }
    ]
  }
}

@sys.description('Client Id of the managed identity.')
output clientId string = managedIdentity.outputs.clientId
