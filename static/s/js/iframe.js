(function() {

  // setup ClipboardJS
  var clipboard = new Clipboard('.js-copy')
    .on('success', function(data) {
      console.log('Copied ' + data.text.length + ' chars.')
    })
    .on('error', function(err) {
      console.log('Err:', err)
    })

}())
