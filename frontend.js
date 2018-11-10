
function komment_form(in_config)
{
  var id = "komment_form_" + in_config.komment_id
  var id_comments = "komment_comments_" + in_config.komment_id

  if ($('#'+id).length == 0)
  {
    document.write('<div id="'+id+'"></div>')
  }

  $('#'+id).load("komment_form.html", null, function(){
    $('#'+id).ajaxForm({
      resetForm: true,
      data: { komment_id: in_config.komment_id },
      dataType: 'json',
      success: function() {
        komment_comments(in_config)
        komment_count(in_config)
      },
      error: function() {
        alert("Error!")
      }
    });
  })

}


function komment_comments(in_config)
{
  var id = "komment_comments_" + in_config.komment_id

  if ($('#'+id).length == 0)
  {
    document.write('<div id="'+id+'"/></div>')
  }

  $('#'+id).load(
    "backend.cgi",
    { "r": "l", "komment_id": in_config.komment_id }
  )

}


function komment_count(in_config)
{
  var id = "komment_count_" + in_config.komment_id

  if ($('#'+id).length == 0)
  {
    document.write('<div id="'+id+'"/></div>')
  }

  $('#'+id).load(
    "backend.cgi",
    { "r": "c", "komment_id": in_config.komment_id }
  )

}

function komment_edit_prepare(in_root)
{
  komment_block = $(in_root).closest(".komment_block")
  komment_message = komment_block.find(".komment_message")
  komment_edit = komment_block.find(".komment_edit")

  komment_edit.css("display", "block")
  komment_message.css("display", "none")
}

function komment_edit_unprepare(in_root)
{
  komment_block = $(in_root).closest(".komment_block")
  komment_message = komment_block.find(".komment_message")
  komment_edit = komment_block.find(".komment_edit")

  komment_message.css("display", "block")
  komment_edit.css("display", "none")
}

function komment_edit_send(in_root)
{
  $(in_root).ajaxSubmit({
    data: { "r": "e", "komment_id": "1" },
    dataType: 'html',
    success: function(r, s, x, form) {
      komment_edit_unprepare(form.get())
      komment_comments({komment_id: 1})
    },
    error: function() {
      alert("Error!")
    }
  });

  return false
}
