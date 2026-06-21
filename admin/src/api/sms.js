import request from '@/utils/request'

export function createSmsTemplate(data) {
  return request({
    url: '/sms/templates',
    method: 'post',
    data
  })
}

export function updateSmsTemplate(id, data) {
  return request({
    url: `/sms/templates/${id}`,
    method: 'put',
    data
  })
}

export function deleteSmsTemplate(id) {
  return request({
    url: `/sms/templates/${id}`,
    method: 'delete'
  })
}

export function getSmsTemplate(id) {
  return request({
    url: `/sms/templates/${id}`,
    method: 'get'
  })
}

export function getSmsTemplateList(params) {
  return request({
    url: '/sms/templates',
    method: 'get',
    params
  })
}

export function reviewSmsTemplate(id, data) {
  return request({
    url: `/sms/templates/${id}/review`,
    method: 'post',
    data
  })
}

export function getActiveTemplates(storeId, templateType) {
  return request({
    url: '/sms/templates/active/list',
    method: 'get',
    params: { store_id: storeId, template_type: templateType }
  })
}

export function createSmsTask(data) {
  return request({
    url: '/sms/tasks',
    method: 'post',
    data
  })
}

export function updateSmsTask(id, data) {
  return request({
    url: `/sms/tasks/${id}`,
    method: 'put',
    data
  })
}

export function deleteSmsTask(id) {
  return request({
    url: `/sms/tasks/${id}`,
    method: 'delete'
  })
}

export function getSmsTask(id) {
  return request({
    url: `/sms/tasks/${id}`,
    method: 'get'
  })
}

export function getSmsTaskList(params) {
  return request({
    url: '/sms/tasks',
    method: 'get',
    params
  })
}

export function startSmsTask(id) {
  return request({
    url: `/sms/tasks/${id}/start`,
    method: 'post'
  })
}

export function pauseSmsTask(id) {
  return request({
    url: `/sms/tasks/${id}/pause`,
    method: 'post'
  })
}

export function getSmsTaskStatistics(params) {
  if (typeof params === 'number' || typeof params === 'string') {
    return request({
      url: `/sms/tasks/${params}/statistics`,
      method: 'get'
    })
  }
  return request({
    url: '/sms/statistics',
    method: 'get',
    params
  })
}

export function calculateTargetCount(data) {
  return request({
    url: '/sms/tasks/target-count',
    method: 'post',
    data
  })
}

export function getSmsRecordList(params) {
  return request({
    url: '/sms/records',
    method: 'get',
    params
  })
}

export function getSmsRecordDetail(id) {
  return request({
    url: `/sms/records/${id}`,
    method: 'get'
  })
}

export function sendTestSms(data) {
  return request({
    url: '/sms/test-send',
    method: 'post',
    data
  })
}
