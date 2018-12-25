import request from '@/utils/request'

export function scanKeys(serverID, data) {
  if (!data.num) {
    data.num = 1000
  }
  return request({
    url: '/server/scan/' + serverID,
    method: 'post',
    data: data
  })
}

export function getKey(serverID, key) {
  return request({
    url: '/server/get/' + serverID,
    method: 'post',
    data: { key: key }
  })
}

export function addKey(serverID, data) {
  return request({
    url: '/server/add/' + serverID,
    method: 'post',
    data: data
  })
}

export function saveKey(serverID, data) {
  return request({
    url: '/server/save/' + serverID,
    method: 'post',
    data: data
  })
}

export function saveRow(serverID, data) {
  return request({
    url: '/server/saveRow/' + serverID,
    method: 'post',
    data: data
  })
}

export function setTTL(serverID, key, ttl) {
  const data = {
    key: key,
    ttl: parseInt(ttl)
  }
  return request({
    url: '/server/expire/' + serverID,
    method: 'post',
    data: data
  })
}

export function renameKey(serverID, key, newKey) {
  const data = {
    old_key: key,
    new_key: newKey
  }
  return request({
    url: '/server/rename/' + serverID,
    method: 'post',
    data: data
  })
}

export function delKey(serverID, key) {
  const data = {
    key: key
  }
  return request({
    url: '/server/del/' + serverID,
    method: 'post',
    data: data
  })
}

export function delRow(serverID, data) {
  return request({
    url: '/server/del-row/' + serverID,
    method: 'post',
    data: data
  })
}

export function scanVal(serverID, data) {
  return request({
    url: '/server/scan-value/' + serverID,
    method: 'post',
    data: data
  })
}

export function delBackground(serverID, keys) {
  const data = {
    keys
  }
  return request({
    url: '/server/del-background/' + serverID,
    method: 'post',
    data: data
  })
}
