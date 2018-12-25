import request from '@/utils/request'

export function getServerList(username, password) {
  return request({
    url: '/server/list',
    method: 'get'
  })
}

export function getDbSize(serverID) {
  return request({
    url: '/server/dbSize/' + serverID,
    method: 'get'
  })
}

export function getServerInfo(serverID) {
  return request({
    url: '/server/info/' + serverID,
    method: 'get'
  })
}

export function stopDelTask(serverID, taskID) {
  const data = {
    task_id: taskID
  }
  return request({
    url: '/server/stop-del-task/' + serverID,
    method: 'post',
    data: data
  })
}

export function removeDelTask(serverID, taskID) {
  const data = {
    task_id: taskID
  }
  return request({
    url: '/server/remove-del-task/' + serverID,
    method: 'post',
    data: data
  })
}

export function refreshKeys(serverID) {
  return request({
    url: '/server/refresh-keys/' + serverID,
    method: 'post',
    data: {}
  })
}
