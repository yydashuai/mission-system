const titleCase = (value) => {
  if (!value) return '--'
  const text = String(value).toLowerCase()
  return text.charAt(0).toUpperCase() + text.slice(1)
}

const missionTypeLabel = (value) => {
  const text = String(value || '').toLowerCase()
  if (text === 'isr') return 'ISR'
  if (!text) return '--'
  return titleCase(text)
}

const priorityLabel = (value) => {
  const text = String(value || '').toLowerCase()
  if (text === 'medium') return 'Normal'
  if (text === 'critical') return 'Critical'
  if (!text) return '--'
  return titleCase(text)
}

const stageTypeLabel = (value) => {
  if (!value) return '--'
  return titleCase(value)
}

const phaseLabel = (value) => {
  if (!value) return '--'
  return titleCase(value)
}

const formatTime = (value) => {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return date.toISOString().slice(11, 16)
}

const parseQuantity = (value) => {
  if (value === undefined || value === null || value === '') return null
  const text = String(value).trim()
  const match = text.match(/^([0-9.]+)([a-zA-Z]+)?$/)
  if (!match) return null
  const number = Number(match[1])
  if (!Number.isFinite(number)) return null
  return { number, unit: match[2] || '' }
}

const parseCpuCores = (value) => {
  const parsed = parseQuantity(value)
  if (!parsed) return null
  const scale = {
    n: 1e-9,
    u: 1e-6,
    m: 1e-3,
    '': 1,
    k: 1e3,
    M: 1e6,
    G: 1e9,
    T: 1e12,
    P: 1e15,
    E: 1e18,
  }
  if (!Object.prototype.hasOwnProperty.call(scale, parsed.unit)) return null
  return parsed.number * scale[parsed.unit]
}

const parseMemoryBytes = (value) => {
  const parsed = parseQuantity(value)
  if (!parsed) return null
  const scale = {
    '': 1,
    Ki: 1024,
    Mi: 1024 ** 2,
    Gi: 1024 ** 3,
    Ti: 1024 ** 4,
    Pi: 1024 ** 5,
    Ei: 1024 ** 6,
    K: 1e3,
    M: 1e6,
    G: 1e9,
    T: 1e12,
    P: 1e15,
    E: 1e18,
  }
  if (!Object.prototype.hasOwnProperty.call(scale, parsed.unit)) return null
  return parsed.number * scale[parsed.unit]
}

const toPercent = (used, capacity) => {
  if (!Number.isFinite(used) || !Number.isFinite(capacity) || capacity <= 0) return 0
  return Math.min(100, Math.round((used / capacity) * 100))
}

const metaValue = (item, key) => (
  item?.metadata?.labels?.[key] || item?.metadata?.annotations?.[key] || ''
)

const isK8sItem = (item) => Boolean(item?.metadata && item?.spec)

const normalizeList = (payload) => {
  if (Array.isArray(payload)) return payload
  if (payload && Array.isArray(payload.items)) return payload.items
  return []
}

const buildMetricsIndex = (payload) => {
  const list = normalizeList(payload)
  const map = new Map()
  list.forEach((item) => {
    const name = item?.metadata?.name
    if (!name) return
    map.set(name, item.usage || {})
  })
  return map
}

const buildPodUsageIndex = (payload) => {
  if (!payload) return { map: new Map(), available: false }
  const list = normalizeList(payload)
  const map = new Map()
  list.forEach((item) => {
    const nodeName = item?.spec?.nodeName
    if (!nodeName) return
    const phase = item?.status?.phase || ''
    if (phase !== 'Running') return
    map.set(nodeName, (map.get(nodeName) || 0) + 1)
  })
  return { map, available: true }
}

const buildObjective = (objective) => {
  if (!objective) return '--'
  if (objective.targetDescription) return objective.targetDescription
  if (objective.targetArea) return objective.targetArea
  if (objective.targetCoordinates) {
    const { latitude, longitude } = objective.targetCoordinates
    if (latitude || longitude) return `${latitude || '--'}, ${longitude || '--'}`
  }
  return '--'
}

const buildFailurePolicy = (config) => {
  const action = config?.failurePolicy?.stageFailureAction
  return action ? titleCase(action) : '--'
}

const buildMissionStages = (specStages, summary) => {
  const summaryMap = new Map()
  if (Array.isArray(summary)) {
    summary.forEach((item) => {
      summaryMap.set(item.name, item.phase)
    })
  }

  return (specStages || []).map((stage) => ({
    name: stage.displayName || stage.name || '--',
    mode: stageTypeLabel(stage.type),
    status: phaseLabel(summaryMap.get(stage.name) || 'Pending'),
    tasks: Array.isArray(stage.flightTasks) ? stage.flightTasks.length : 0,
    dependsOn: Array.isArray(stage.dependsOn) ? stage.dependsOn : [],
  }))
}

const buildStageTasks = (taskStatus, taskSpec) => {
  if (Array.isArray(taskStatus) && taskStatus.length) {
    return taskStatus.map((item) => ({
      name: item.name || '--',
      status: phaseLabel(item.phase),
      eta: '--',
      node: item.aircraftNode || '--',
    }))
  }

  if (Array.isArray(taskSpec) && taskSpec.length) {
    return taskSpec.map((item) => ({
      name: item.name || '--',
      status: 'Pending',
      eta: '--',
      node: '--',
    }))
  }

  return []
}

const buildTaskConstraints = (requirement) => {
  if (!requirement) return []
  const out = []
  if (requirement.type) out.push(`type:${requirement.type}`)
  if (requirement.requiredHardpoints) out.push(`hardpoint:${requirement.requiredHardpoints}+`)
  if (requirement.minFuelLevel) out.push(`fuel>=${requirement.minFuelLevel}%`)
  if (requirement.preferredLocation) out.push(`zone:${requirement.preferredLocation}`)
  if (Array.isArray(requirement.capabilities)) {
    requirement.capabilities.forEach((cap) => {
      if (cap) out.push(`cap:${cap}`)
    })
  }
  return out
}

const buildTaskConditions = (conditions) => {
  if (!Array.isArray(conditions)) return []
  return conditions.map((item) => {
    const type = item.type || '--'
    const status = String(item.status || '').toLowerCase()
    const reason = item.reason || ''
    const message = item.message || ''
    const detail = reason && message ? `${reason}: ${message}` : (message || reason || '--')

    let tone = 'warn'
    if (status === 'true') tone = 'ok'
    if (status === 'false' && (reason.toLowerCase().includes('fail') || type.toLowerCase().includes('fail'))) {
      tone = 'err'
    }

    return { label: type, tone, detail }
  })
}

const buildWeaponResources = (resources) => {
  if (!resources) return '--'
  const parts = []
  if (Number.isFinite(resources.hardpoints)) parts.push(`hp ${resources.hardpoints}`)
  if (Number.isFinite(resources.weight)) parts.push(`weight ${resources.weight}`)
  if (Number.isFinite(resources.power)) parts.push(`power ${resources.power}`)
  if (resources.cooling) parts.push(`cooling ${resources.cooling}`)
  return parts.length ? parts.join(' · ') : '--'
}

const buildWeaponUsage = (usage) => {
  if (!usage) return '--'
  if (Number.isFinite(usage.totalDeployed)) return `${usage.totalDeployed} deployed`
  if (Number.isFinite(usage.totalFired)) return `${usage.totalFired} fired`
  return '--'
}

const buildWeaponImage = (image) => {
  if (!image) return '--'
  if (image.repository && image.tag) return `${image.repository}:${image.tag}`
  return image.repository || '--'
}

const buildNodeRole = (labels) => {
  if (!labels) return 'worker'
  if (Object.prototype.hasOwnProperty.call(labels, 'node-role.kubernetes.io/control-plane')) {
    return 'control-plane'
  }
  if (Object.prototype.hasOwnProperty.call(labels, 'node-role.kubernetes.io/master')) {
    return 'master'
  }
  if (Object.prototype.hasOwnProperty.call(labels, 'node-role.kubernetes.io/worker')) {
    return 'worker'
  }
  return 'worker'
}

const buildNodeZone = (labels) => (
  labels?.['topology.kubernetes.io/zone'] || labels?.['failure-domain.beta.kubernetes.io/zone'] || '--'
)

const buildNodeStatus = (conditions) => {
  if (!Array.isArray(conditions)) return 'Unknown'
  const ready = conditions.find((item) => item.type === 'Ready')
  if (!ready) return 'Unknown'
  return ready.status === 'True' ? 'Ready' : 'NotReady'
}

const buildEventTone = (value, reason) => {
  const text = String(value || '').toLowerCase()
  if (text === 'normal') return 'ok'
  if (text === 'warning') return 'warn'
  if (text === 'error') return 'err'
  if (String(reason || '').toLowerCase().includes('fail')) return 'err'
  return 'muted'
}

export const normalizeMissionList = (payload) => {
  const list = normalizeList(payload)
  if (!list.length) return []
  if (!isK8sItem(list[0])) return list

  return list.map((item) => {
    const spec = item.spec || {}
    const status = item.status || {}
    const tasksCount = status.statistics?.totalFlightTasks
      ?? (spec.stages || []).reduce((sum, stage) => sum + (stage.flightTasks?.length || 0), 0)

    return {
      name: spec.missionName || item.metadata?.name || '--',
      type: missionTypeLabel(spec.missionType),
      priority: priorityLabel(spec.priority),
      status: phaseLabel(status.phase),
      commander: metaValue(item, 'commander') || metaValue(item, 'mission.airforce.mil/commander') || '--',
      region: spec.objective?.targetArea || metaValue(item, 'region') || '--',
      updated: formatTime(status.lastUpdateTime || item.metadata?.creationTimestamp),
      objective: buildObjective(spec.objective),
      failurePolicy: buildFailurePolicy(spec.config),
      tasks: tasksCount,
      stages: buildMissionStages(spec.stages, status.stagesSummary),
    }
  })
}

export const normalizeStageList = (payload) => {
  const list = normalizeList(payload)
  if (!list.length) return []
  if (!isK8sItem(list[0])) return list

  return list.map((item) => {
    const spec = item.spec || {}
    const status = item.status || {}
    return {
      name: spec.stageName || metaValue(item, 'stage-name') || item.metadata?.name || '--',
      mission: spec.missionRef?.name || metaValue(item, 'mission') || '--',
      index: spec.stageIndex || Number(metaValue(item, 'stage-index')) || '--',
      mode: stageTypeLabel(spec.stageType),
      status: phaseLabel(status.phase),
      timeout: spec.config?.timeout || '--',
      dependsOn: Array.isArray(spec.dependsOn) ? spec.dependsOn : [],
      tasks: buildStageTasks(status.flightTasksStatus, spec.flightTasks),
    }
  })
}

export const normalizeFlightTaskList = (payload) => {
  const list = normalizeList(payload)
  if (!list.length) return []
  if (!isK8sItem(list[0])) return list

  return list.map((item) => {
    const spec = item.spec || {}
    const status = item.status || {}
    const weaponNames = Array.isArray(spec.weaponLoadout)
      ? spec.weaponLoadout.map((load) => load.weaponRef?.name).filter(Boolean)
      : []

    return {
      name: item.metadata?.name || '--',
      stage: spec.stageRef?.name || metaValue(item, 'stage') || metaValue(item, 'stageRef') || '--',
      mission: metaValue(item, 'mission') || '--',
      status: phaseLabel(status.phase),
      pod: status.podRef?.name || '--',
      node: status.schedulingInfo?.assignedNode || '--',
      weapon: weaponNames[0] || '--',
      attempts: status.schedulingInfo?.schedulingAttempts ?? 0,
      scheduledAt: formatTime(status.schedulingInfo?.assignedTime),
      conditions: buildTaskConditions(status.conditions),
      constraints: buildTaskConstraints(spec.aircraftRequirement),
      podStatus: status.executionStatus?.currentPhase || phaseLabel(status.phase),
      sidecars: weaponNames.map((name) => `weapon-${name}`),
    }
  })
}

export const normalizeWeaponList = (payload) => {
  const list = normalizeList(payload)
  if (!list.length) return []
  if (!isK8sItem(list[0])) return list

  return list.map((item) => {
    const spec = item.spec || {}
    const status = item.status || {}
    return {
      name: spec.weaponName || item.metadata?.name || '--',
      status: phaseLabel(status.phase),
      image: buildWeaponImage(spec.image),
      version: spec.version?.current || '--',
      usage: buildWeaponUsage(status.usage),
      aircraft: spec.compatibility?.aircraftTypes || [],
      hardpoints: spec.compatibility?.hardpointTypes || [],
      resources: buildWeaponResources(spec.resources),
    }
  })
}

export const normalizeNodeList = (payload, metricsPayload = null, podsPayload = null) => {
  const list = normalizeList(payload)
  if (!list.length) return []
  if (!list[0]?.metadata) return list

  const metricsIndex = buildMetricsIndex(metricsPayload)
  const podUsage = buildPodUsageIndex(podsPayload)

  return list.map((item) => {
    const labels = item.metadata?.labels || {}
    const capacity = item.status?.capacity || {}
    const allocatable = item.status?.allocatable || {}
    const podsCap = allocatable.pods || capacity.pods || '--'
    const name = item.metadata?.name || '--'
    const usage = metricsIndex.get(name) || {}
    const cpuUsed = parseCpuCores(usage.cpu)
    const cpuCap = parseCpuCores(allocatable.cpu || capacity.cpu)
    const memoryUsed = parseMemoryBytes(usage.memory)
    const memoryCap = parseMemoryBytes(allocatable.memory || capacity.memory)
    const podsUsed = podUsage.map.get(name)
    const podsUsedLabel = podUsage.available ? (podsUsed ?? 0) : '--'

    return {
      name,
      role: buildNodeRole(labels),
      status: buildNodeStatus(item.status?.conditions),
      cpu: toPercent(cpuUsed, cpuCap),
      memory: toPercent(memoryUsed, memoryCap),
      pods: `${podsUsedLabel} / ${podsCap}`,
      zone: buildNodeZone(labels),
    }
  })
}

export const normalizeEventList = (payload) => {
  const list = normalizeList(payload)
  if (!list.length) return []

  return list.map((item) => {
    const time = item.eventTime || item.lastTimestamp || item.firstTimestamp || item.metadata?.creationTimestamp
    const type = item.type || item.type?.type || item.reason || ''
    const reason = item.reason || ''
    const message = item.message || item.note || ''
    const label = reason && message ? `${reason}: ${message}` : (message || reason || '--')
    const scope = item.source?.component || item.reportingController || item.metadata?.namespace || '--'
    return {
      time: formatTime(time),
      scope,
      level: buildEventTone(type, reason),
      message: label,
    }
  })
}
