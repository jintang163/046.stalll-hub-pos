const { app, BrowserWindow, ipcMain, Notification } = require('electron')
const path = require('path')
const Database = require('./database')
const NSQClient = require('./nsq')
const IPC = require('./ipc')
const Store = require('electron-store')

let mainWindow = null
let db = null
let nsq = null
const store = new Store({
  name: 'pos-config',
  defaults: {
    storeID: 1,
    apiBaseURL: 'http://localhost:8080/api/v1',
    nsqLookupd: 'http://localhost:4161',
    nsqd: 'localhost:4150'
  }
})

function createWindow() {
  const isDev = process.env.NODE_ENV === 'development'
  
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 1200,
    minHeight: 768,
    title: '大排档收银系统',
    autoHideMenuBar: true,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      enableRemoteModule: false
    }
  })

  if (isDev) {
    mainWindow.loadURL('http://localhost:5174')
    mainWindow.webContents.openDevTools()
  } else {
    mainWindow.loadFile(path.join(__dirname, '../dist/index.html'))
  }

  mainWindow.on('closed', () => {
    mainWindow = null
  })
}

function initDatabase() {
  const userDataPath = app.getPath('userData')
  const dbPath = path.join(userDataPath, 'stalll-pos', 'cashier.db')
  db = new Database(dbPath)
  db.init()
  console.log('[Main] 数据库初始化完成:', dbPath)
}

function initNSQ() {
  const nsqConfig = store.get('nsqd') || 'localhost:4150'
  const [nsqdHost, nsqdPortStr] = nsqConfig.split(':')
  const nsqdPort = parseInt(nsqdPortStr) || 4150
  const lookupdHTTPAddresses = [store.get('nsqLookupd') || 'http://localhost:4161']
  const storeID = store.get('storeID') || 1

  nsq = new NSQClient({
    nsqdHost,
    nsqdPort,
    lookupdHTTPAddresses,
    storeID
  })

  nsq.setDatabase(db)
  nsq.setMainWindow(mainWindow)

  nsq.on('connected', () => {
    console.log('[Main] NSQ连接成功')
    if (mainWindow && !mainWindow.isDestroyed()) {
      mainWindow.webContents.send('nsq:status', { connected: true })
    }
  })

  nsq.on('closed', () => {
    console.log('[Main] NSQ连接已关闭')
    if (mainWindow && !mainWindow.isDestroyed()) {
      mainWindow.webContents.send('nsq:status', { connected: false })
    }
  })

  nsq.on('message', ({ type, payload, messageId }) => {
    console.log('[Main] 收到NSQ消息:', type, messageId)
  })

  nsq.connect()
  console.log('[Main] NSQ初始化完成')
}

app.whenReady().then(() => {
  initDatabase()
  
  IPC.init(ipcMain, db, mainWindow, nsq, store, app)
  
  createWindow()

  if (mainWindow) {
    initNSQ()
  }

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on('window-all-closed', () => {
  if (nsq) {
    nsq.close()
    nsq = null
  }
  if (db) {
    db.close()
    db = null
  }
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('before-quit', (event) => {
  if (nsq) {
    nsq.close()
  }
  if (db) {
    db.close()
  }
})

app.on('ready', () => {
  if (Notification.isSupported()) {
    console.log('[Main] 通知功能已启用')
  }
})
