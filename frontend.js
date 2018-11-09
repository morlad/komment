
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

  jQuery.getJSON(
    "backend.cgi", 
    { "r": "c", "komment_id": in_config.komment_id },
    function(in_json) {
      $('#'+id).html("Count = "+in_json.count)
    }
  )

}

