
function komment_sanitize_komment_id(in_kid)
{
  in_kid = '' + in_kid
  var rex = /[^a-zA-Z0-9_]+/g;
  return in_kid.replace(rex, "-").toLowerCase()
}

function komment_form(in_config)
{
  var komment_id = komment_sanitize_komment_id(in_config.komment_id)
  var id = "komment_form_" + komment_id
  var id_comments = "komment_comments_" + komment_id

  if ($('#'+id).length == 0)
  {
    document.write('<div id="'+id+'"></div>')
  }

  $('#'+id).load("form.html", null, function(){
    $('#'+id).ajaxForm({
      resetForm: true,
      data: { "r": "a", komment_id: in_config.komment_id },
      dataType: 'json',
      success: function() {
        komment_comments(in_config, true)
        komment_count(in_config, true)
      },
      error: function() {
        alert("Error!")
      }
    });
  })

}


function komment_comments(in_config, in_do_not_add)
{
  var komment_id = komment_sanitize_komment_id(in_config.komment_id)
  var id = "komment_comments_" + komment_id

  if ($('#'+id).length == 0)
  {
    if (in_do_not_add) { return }
    document.write('<div id="'+id+'"/></div>')
  }

  $('#'+id).load(
    "komment.cgi",
    { "r": "l", "komment_id": in_config.komment_id }
  )

}


function komment_count(in_config, in_do_not_add)
{
  var komment_id = komment_sanitize_komment_id(in_config.komment_id)
  var id = "komment_count_" + komment_id

  if ($('#'+id).length == 0)
  {
    if (in_do_not_add) { return }
    document.write('<div id="'+id+'"/></div>')
  }

  $('#'+id).load(
    "komment.cgi",
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

function komment_edit_send(in_root, in_komment_id)
{
  var komment_id = komment_sanitize_komment_id(in_komment_id)
  $(in_root).ajaxSubmit({
    data: { "r": "e" },
    dataType: 'html',
    success: function(r, s, x, form) {
      komment_edit_unprepare(form.get())
      komment_comments({komment_id: in_komment_id})
    },
    error: function() {
      alert("Error!")
    }
  });

  return false
}
