function Page () {
  return {
    sidebarButton () {
      return document.getElementById('top-nav-title')
    },
    addButton () {
      return document.getElementById('add')
    },
    editor () {
      return document.getElementById('editor')
    },
    files () {
      return document.getElementById('files')
    },
    file () {
      return document.createElement('li')
    },
    sidebar () {
      return document.getElementById('sidebar')
    },
    deleteButton () {
      return document.getElementById('delete-button')
    },
    downloadButton () {
      return document.getElementById('download-button')
    }
  }
}
