window.addEventListener("load", ()=>{
    const input    = document.getElementById('file-input')
    const preview  = document.getElementById('file-preview')
    const dropZone = document.getElementById('drop-zone')
    const form     = document.getElementById('upload-form')
    const btn      = document.getElementById('upload-btn')
    const progressWrap = document.getElementById('progress-wrap')
    const fill     = document.getElementById('progress-fill')
    const label    = document.getElementById('progress-label')

    input.addEventListener('change', () => {
      if (input.files[0]) preview.textContent = input.files[0].name
    })

    // Drag & drop
    ['dragenter','dragover'].forEach(evt => {
      dropZone.addEventListener(evt, e => { e.preventDefault(); dropZone.classList.add('drag-over');})
    })
    ['dragleave','drop'].forEach(evt => {
      dropZone.addEventListener(evt, e => { e.preventDefault(); dropZone.classList.remove('drag-over') })
    })
    dropZone.addEventListener('drop', e => {
      const files = e.dataTransfer.files
      if (files.length) {
        input.files = files
        preview.textContent = files[0].name
      }
    });

    // Upload com barra de progresso via XHR
    form.addEventListener('submit', e => {
      e.preventDefault()
      if (!input.files[0]) return

      const data = new FormData(form)
      const xhr  = new XMLHttpRequest()

      progressWrap.classList.add('active')
      btn.disabled = true
      btn.textContent = 'Enviando…'

      xhr.upload.addEventListener('progress', ev => {
        if (ev.lengthComputable) {
          const pct = Math.round(ev.loaded / ev.total * 100)
          fill.style.width  = pct + '%'
          label.textContent = pct + '%'
        }
      });

    xhr.addEventListener('load', () => { window.location = xhr.responseURL || '/'})
    xhr.addEventListener('error', () => {
      btn.disabled = false
      btn.textContent = 'Tentar novamente'
      label.textContent = 'Erro no envio'
    });

    xhr.open('POST', '/upload')
    xhr.send(data)
  });
})