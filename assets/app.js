//(function($) {
  $(document).ready(function() {
    var linesToFetch = 2;
    var needle = $('section.results').data('needle');
    $('section.results').highlight(needle);

    // compares 2 objects (e.g. in a sort call) using the specified key
    function cmp_key(key) {
      return function(a,b) {
        a = a[key], b = b[key]
        if (a < b)
          return -1;
        else if (a > b)
          return 1;
        else
          return 0;
      }
    }

    function getAndInsertContext(messageId, linesToFetch, direction) {
      console.log(this);
      $.ajax({
        url: "/context/",
        data: {messageId: messageId,
               linesToFetch: linesToFetch,
               direction: direction},
        context: this,
      }).done(function(messages) {
        // sort messages ASC if aftercontext, DESC if beforecontext
        messages = messages.sort(cmp_key("MessageId"));
        if (direction == -1)
          messages.reverse()

        var msgElem, msg;
        var directionSelect = (direction == -1) ? ".before" : ".after";
        var directionFunc = (direction == -1) ? "prependTo" : "appendTo";
        for (var i = 0; i < messages.length; i++) {
          msg = messages[i];
          msgElem = $(this).children('.message').first().clone()
          //$(msgElem).data('messageid', msg.MessageId);
          $(msgElem).attr('data-messageid', msg.MessageId);
          $(msgElem).children('.channel').text(msg.Channel);
          $(msgElem).children('.time').text(msg.Time);
          $(msgElem).children('.sender').text(msg.Sender);
          $(msgElem).children('.text').text(msg.Text);
          $(msgElem)[directionFunc]($(this).parent().children(directionSelect + '.context'));
        }
        $(this).highlight(needle);
      });
    }

    $('.matching-line').click(function() {
      var lowMsgId = $(this).siblings('.before.context').children().first().data('messageid') ||
                     $(this).children('.message').data('messageid');
      var highMsgId = $(this).siblings('.after.context').children().last().data('messageid') ||
                      $(this).children('.message').data('messageid');
      getAndInsertContext.call(this, lowMsgId, linesToFetch, -1);
      getAndInsertContext.call(this, highMsgId, linesToFetch, 1);
    });
    $('.before.context').click(function() {
      var messageId = $(this).children().first().data('messageid');
      getAndInsertContext.call(this, messageId, linesToFetch, -1);
    });
    $('.after.context').click(function() {
      var messageId = $(this).children().last().data('messageid');
      getAndInsertContext.call(this, messageId, linesToFetch, 1);
    });
  });
//})(jQuery);
