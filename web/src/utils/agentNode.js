export function agentCapabilities(agent) {
  if (Array.isArray(agent?.capabilities)) return agent.capabilities
  try {
    const values = JSON.parse(agent?.capabilities || '[]')
    return Array.isArray(values) ? values : []
  } catch {
    return []
  }
}

export function isDockerAgent(agent) {
  return agentCapabilities(agent).includes('runtime:docker')
}

export function agentDockerImage(agent) {
  const capability = agentCapabilities(agent).find(item => String(item).startsWith('docker_image:'))
  return capability ? capability.slice('docker_image:'.length) : ''
}

export function agentDeploymentType(agent) {
  return isDockerAgent(agent) ? 'docker' : 'agent'
}

export function agentTypeText(agent) {
  return isDockerAgent(agent) ? 'Docker' : 'Agent'
}
