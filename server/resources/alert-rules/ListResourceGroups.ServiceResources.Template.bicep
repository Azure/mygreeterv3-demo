// TODO: Should we create alert rules for each region/location? Should we create a workspace for every region?
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
resource logAnalyticsWorkspace 'Microsoft.OperationalInsights/workspaces@2022-10-01' existing = {
  name: 'servicehubval-${resourcesName}-workspace'
  scope: resourceGroup(subscriptionId, resourceGroupName)
}

module qpsAlertRule 'br/public:avm/res/insights/scheduled-query-rule:0.1.2' = {
  name: 'mygreeterv3-ListResourceGroups-query-per-second'
  scope: resourceGroup(subscriptionId, resourceGroupName)
  params: {
    name: 'mygreeterv3-ListResourceGroups-query-per-second'
    location: location
    alertDescription: 'This is the alert for mygreeterv3 for the calculated metric query per second.'
    criterias: {
      allOf: [
        {
          metricMeasureColumn:'QPS'
          operator: 'GreaterThan'
          query:'let method = "ListResourceGroups";\nlet binSizeMinute = timespan(5m);\nlet binSizeSecond = binSizeMinute / 1s;\nContainerLogV2\n| where ContainerName == "servicehubval-mygreeterv3-server"\n| where LogMessage["component"] == "server"\n| where LogMessage["method"] == method or isempty(method)\n| where LogMessage["msg"] == "finished call"\n| summarize QPS = count()/binSizeSecond by tostring(LogMessage["code"]), bin(todatetime(LogMessage["time"]), binSizeMinute)\n\n'
          threshold: '0.05'
          timeAggregation: 'Maximum'
        }
      ]
    }
    enabled: true
    windowSize: 'PT5M'
    evaluationFrequency: 'PT5M'
    severity: 4
    autoMitigate: true
    scopes: [
      logAnalyticsWorkspace.id
    ]
  }
}

module errorRatioAlertRule 'br/public:avm/res/insights/scheduled-query-rule:0.1.2' = {
  name: 'mygreeterv3-ListResourceGroups-error-ratio'
  scope: resourceGroup(subscriptionId, resourceGroupName)
  params: {
    name: 'mygreeterv3-ListResourceGroups-error-ratio'
    location: location
    alertDescription: 'This is the alert for mygreeterv3 for the calculated metric error ratio.'
    criterias: {
      allOf: [
        {
          metricMeasureColumn:'ratio'
          operator: 'GreaterThan'
          query: 'let method = "ListResourceGroups";\nContainerLogV2\n| where ContainerName == "servicehubval-mygreeterv3-server"\n| where LogMessage["component"] == "server"\n| where LogMessage["method"] == method\n| where LogMessage["msg"] == "finished call"\n| extend timePoint = todatetime(LogMessage["time"])\n| extend code = tostring(LogMessage["code"])\n| summarize \n total = count(),\nerror = countif(code != "OK") \nby bin(timePoint,5m)\n| extend ratio = round(error * 100.0/total, 3)\n| project timePoint, ratio\n\n'
          threshold: '0.05'
          timeAggregation: 'Maximum'
        }
      ]
    }
    enabled: true
    windowSize: 'PT5M'
    evaluationFrequency: 'PT5M'
    severity: 4
    autoMitigate: true
    scopes: [
      logAnalyticsWorkspace.id
    ]
  }
}

module latencyAlertRule 'br/public:avm/res/insights/scheduled-query-rule:0.1.2' = {
  name: 'mygreeterv3-ListResourceGroups-latency-by-error-code'
  scope: resourceGroup(subscriptionId, resourceGroupName)
  params: {
    name: 'mygreeterv3-ListResourceGroups-latency-by-error-code'
    location: location
    alertDescription: 'This is the alert for mygreeterv3 for the calculated metric error ratio.'
    criterias: {
      allOf: [
        {
          metricMeasureColumn:'latency'
          operator: 'GreaterThan'
          query: 'let method = "ListResourceGroups";\nlet binSizeMinute = timespan(5m);\nlet binSizeSecond = binSizeMinute / 1s;\nContainerLogV2\n| where ContainerName == "servicehubval-mygreeterv3-server"\n| where LogMessage["component"] == "server"\n| where LogMessage["method"] == method\n| where LogMessage["msg"] == "finished call"\n| summarize latency = avg(todouble(LogMessage["time_ms"]))\n    by\n    tostring(LogMessage["code"]),\n    bin(todatetime(LogMessage["time"]), binSizeMinute)\n\n'
          threshold: '1600'
          timeAggregation: 'Maximum'
        }
      ]
    }
    enabled: true
    windowSize: 'PT5M'
    evaluationFrequency: 'PT5M'
    severity: 4
    autoMitigate: true
    scopes: [
      logAnalyticsWorkspace.id
    ]
  }
}
