import request from '@/utils/request'

export function getMemberList(params) {
  return request({
    url: '/members',
    method: 'get',
    params
  })
}

export function getMember(id) {
  return request({
    url: `/members/${id}`,
    method: 'get'
  })
}

export function createMember(data) {
  return request({
    url: '/members',
    method: 'post',
    data
  })
}

export function updateMember(id, data) {
  return request({
    url: `/members/${id}`,
    method: 'put',
    data
  })
}

export function deleteMember(id) {
  return request({
    url: `/members/${id}`,
    method: 'delete'
  })
}

export function adjustPoints(id, data) {
  return request({
    url: `/members/${id}/points`,
    method: 'post',
    data
  })
}
