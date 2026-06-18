const { app, BrowserWindow, ipcMain, Notification } = require('electron')
const path = require('path')
const Database = require('./database')
const IPC = require('./ipc')

let mainWindow = null
let db = null

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

app.whenReady().then(() => {
  db = new Database(app.getPath('userData'))
  db.init()
  
  IPC.init(ipcMain, db, mainWindow)
  
  createWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on('window-all-closed', () => {
  if (db) {
    db.close()
  }
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('ready', () => {
  if (Notification.isSupported()) {
    console.log('通知功能已启用')
  }
})
