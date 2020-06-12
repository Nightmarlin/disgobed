/*
Package disgobed wraps the discordgo embed with helper functions to facilitate easier construction.
Note that all methods in this module act ByReference, directly changing the embed they are called on, instead of
creating and returning a new embed
*/
package disgobed

import (
	"fmt"
	"time"

	"github.com/andersfylling/disgord"
)

/*
Embed wraps the disgord.Embed type and adds features. Never create it directly, instead use the NewEmbed function

	embed := NewEmbed()

and call the methods to set the properties, allowing for chains that look like this

	embed := NewEmbed()
		.SetTitle(`example`)
		.SetDescription(`test`)
		.SetURL(`example.com`)
		.Finalize()

for healthy embedment!
*/
type Embed struct {
	*disgord.Embed
	Errors *[]error
}

/*
Finalize strips away the extra functions and returns the wrapped type. It should always be called before an embed is
sent. Finalize will also purge the error cache!
*/
func (e *Embed) Finalize() (*disgord.Embed, *[]error) {
	defer func(e *Embed) { e.Errors = nil }(e)
	return e.Embed, e.Errors
}

/*
addError takes a message string and adds it to the error slice stored in Embed. If the pointer is nil a new error slice
is created. This function takes the same inputs as fmt.Sprintf
*/
func (e *Embed) addError(format string, values ...interface{}) {
	if e.Errors == nil {
		e.Errors = &[]error{}
	}
	*e.Errors = append(*e.Errors, fmt.Errorf(format, values...))
}

/*
addRawError takes a pre-existing error and adds it to the stored slice. If the pointer is nil a new error slice is
created.
*/
func (e *Embed) addRawError(err error) {
	if e.Errors == nil {
		e.Errors = &[]error{}
	}
	*e.Errors = append(*e.Errors, err)
}

/*
addAllRawErrors takes a pre-existing error slice and adds it to the stored slice. If the pointer is nil a new error
slice is created.
*/
func (e *Embed) addAllRawErrors(errs *[]error) {
	if errs == nil {
		return
	}
	for _, err := range *errs {
		e.addRawError(err)
	}
}

/*
NewEmbed creates and returns an empty embed
*/
func NewEmbed() *Embed {
	res := &Embed{
		Embed:  &disgord.Embed{},
		Errors: nil,
	}
	return res
}

/*
SetTitle edits the embed's title and returns the pointer to the embed. The discord API limits embed titles to 256
characters, so this function will do nothing if len(title) > 256
(This function fails silently)
*/
func (e *Embed) SetTitle(title string) *Embed {
	if len(title) <= lowerCharLimit {
		e.Title = title
	} else {
		e.addError(characterCountExceedsLimitErrTemplateString, `embed title`, lowerCharLimit, len(title), title)
	}
	return e
}

/*
SetDescription edits the embed's description and returns the pointer to the embed. The discord API limits embed
descriptions to 2048 characters, so this function will do nothing if len(desc) > 2048
(This function fails silently)
*/
func (e *Embed) SetDescription(desc string) *Embed {
	if len(desc) <= upperCharLimit {
		e.Description = desc
	} else {
		e.addError(characterCountExceedsLimitLongErrTemplateString, `embed description`, upperCharLimit, len(desc))
	}
	return e
}

/*
SetURL edits the embed's main URL and returns the pointer to the embed
*/
func (e *Embed) SetURL(url string) *Embed {
	e.URL = url
	return e
}

/*
SetColor edits the embed's highlight colour and returns the pointer to the embed.
Color values must be between 0 and 16777215 otherwise the change will not be registered
(This function fails silently)
*/
func (e *Embed) SetColor(color int) *Embed {
	if color >= 0 && color < maxColorValue {
		e.Color = color
	} else {
		e.addError(valueNotBetweenErrTemplateString, `embed color`, color, 0, maxColorValue)
	}
	return e
}

/*
SetCurrentTimestamp sets the embed's timestamp to the current UTC time in the appropriate discord format and returns
the pointer to the embed
*/
func (e *Embed) SetCurrentTimestamp() *Embed {
	utcTime := disgord.Time{Time: time.Now().UTC()}
	return e.setRawTimestamp(utcTime)
}

/*
SetCustomTimestamp sets the embed's timestamp to that specified by the time.Time structure passed to it.
The value stored is the corresponding UTC time in the appropriate discord format.
SetCustomTimestamp returns the pointer to the embed
*/
func (e *Embed) SetCustomTimestamp(t time.Time) *Embed {
	utcTime := disgord.Time{Time: t.UTC()}
	return e.setRawTimestamp(utcTime)
}

/*
Sets the timestamp string to the argument and returns the pointer to the embed. Was exposed but the potential for error
was too high, so has since been replaced with SetCustomTimestamp(t time.Time)
*/
func (e *Embed) setRawTimestamp(timestamp disgord.Time) *Embed {
	e.Timestamp = timestamp
	return e
}

/*
InlineAllFields sets the Inline property on all currently attached fields to true and returns the pointer to the embed
*/
func (e *Embed) InlineAllFields() *Embed {
	for _, f := range e.Fields {
		f.Inline = true
	}
	return e
}

/*
OutlineAllFields sets the Inline property on all currently attached fields to false and returns the pointer to the embed
*/
func (e *Embed) OutlineAllFields() *Embed {
	for _, f := range e.Fields {
		f.Inline = false
	}
	return e
}

/*
AddFields takes N Field structures and adds them to the embed, then returns the pointer to the embed.
Note that Field structures are `Finalize`d once added and should not be changed after being added.
The discord API limits embeds to having 25 Fields, so this function will add the first items from the list until that
limit is reached
(This function fails silently)
*/
func (e *Embed) AddFields(fields ...*Field) *Embed {
	for _, f := range fields {
		e.AddField(f)
	}
	return e
}

/*
AddRawFields takes N disgord.EmbedField structures and adds them to the embed, then returns the pointer to the
embed. The discord API limits embeds to having 25 Fields, so this function will add the first items from the list until
that limit is reached
(This function fails silently)
*/
func (e *Embed) AddRawFields(fields ...*disgord.EmbedField) *Embed {
	for _, f := range fields {
		e.AddRawField(f)
	}
	return e
}

/*
AddField takes a Field structure and adds it to the embed, then returns the pointer to the embed.
Note that the Field structure is `Finalize`d once added and should not be changed after being added.
The discord API limits embeds to having 25 Fields, so this function will not add any fields if the limit has already
been reached. All errors are propagated to the main embed
(This function fails silently)
*/
func (e *Embed) AddField(field *Field) *Embed {
	res, errs := field.Finalize()
	e.addAllRawErrors(errs)
	return e.AddRawField(res)
}

/*
AddRawField takes a disgord.EmbedField structure and adds it to the embed, then returns the pointer to the
embed. The discord API limits embeds to having 25 Fields, so this function will not add any fields if the limit has
already been reached
(This function fails silently)
*/
func (e *Embed) AddRawField(field *disgord.EmbedField) *Embed {
	if len(e.Fields) < maxFieldCount {
		e.Fields = append(e.Fields, field)
	} else {
		e.addError(fieldLimitReachedErrTemplateString, field.Name, maxFieldCount)
	}
	return e
}

/*
SetAuthor takes an Author structure and sets the embed's author field to it, then returns the pointer to the embed.
Note that the Author structure is `Finalize`d once added and should not be changed after being added. All errors are
propagated to the main embed
*/
func (e *Embed) SetAuthor(author *Author) *Embed {
	res, errs := author.Finalize()
	e.addAllRawErrors(errs)
	return e.SetRawAuthor(res)
}

/*
SetRawAuthor takes a disgord.EmbedAuthor and sets the embed's author field to it, then returns the pointer to
the embed
*/
func (e *Embed) SetRawAuthor(author *disgord.EmbedAuthor) *Embed {
	e.Author = author
	return e
}

/*
SetThumbnail takes a Thumbnail structure and sets the embed's thumbnail field to it, then returns the pointer to the
embed. Note that the Thumbnail structure is `Finalize`d once added and should not be changed after being added
*/
func (e *Embed) SetThumbnail(thumb *Thumbnail) *Embed {
	res, errs := thumb.Finalize()
	e.addAllRawErrors(errs)
	return e.SetRawThumbnail(res)
}

/*
SetRawThumbnail takes a disgord.EmbedThumbnail and sets the embed's thumbnail field to it, then returns the
pointer to the embed
*/
func (e *Embed) SetRawThumbnail(thumb *disgord.EmbedThumbnail) *Embed {
	e.Thumbnail = thumb
	return e
}

/*
SetProvider allows you to set the provider of an embed. It will then return the pointer to the embed.
See the provider.go docs for some extra information
*/
func (e *Embed) SetProvider(provider *Provider) *Embed {
	res, errs := provider.Finalize()
	if errs != nil { // This should never run
		e.addAllRawErrors(errs)
	}
	return e.SetRawProvider(res)
}

/*
SetRawProvider allows you to set the disgord.EmbedProvider of an embed.
It will then return the pointer to the embed.
See the provider.go docs for some extra information
*/
func (e *Embed) SetRawProvider(provider *disgord.EmbedProvider) *Embed {
	e.Provider = provider
	return e
}

/*
SetFooter sets the embed's footer property to the Footer passed to it, then returns the pointer to the embed.
Note that the Footer structure is `Finalize`d once added and should not be changed after being added. Footer errors
will be propagated into the embed struct
*/
func (e *Embed) SetFooter(footer *Footer) *Embed {
	res, errs := footer.Finalize()
	e.addAllRawErrors(errs)
	return e.SetRawFooter(res)
}

/*
SetRawFooter takes a disgord.EmbedThumbnail and sets the embed's thumbnail field to it, then returns the
pointer to the embed
*/
func (e *Embed) SetRawFooter(footer *disgord.EmbedFooter) *Embed {
	e.Footer = footer
	return e
}

/*
SetVideo sets the embed's video property to the Video passed to it, then returns the pointer to the embed.
Note that the Video structure is `Finalize`d once added and should not be changed after being added
*/
func (e *Embed) SetVideo(vid *Video) *Embed {
	res, errs := vid.Finalize()
	e.addAllRawErrors(errs)
	return e.SetRawVideo(res)
}

/*
SetRawVideo takes a disgord.EmbedVideo and sets the embed's thumbnail field to it, then returns the pointer to
the embed
*/
func (e *Embed) SetRawVideo(vid *disgord.EmbedVideo) *Embed {
	e.Video = vid
	return e
}

/*
SetImage sets the embed's image property to the Image passed to it, then returns the pointer to the embed.
Note that the Image structure is `Finalize`d once added and should not be changed after being added. Image errors
will be propagated into the embed struct
*/
func (e *Embed) SetImage(img *Image) *Embed {
	res, errs := img.Finalize()
	e.addAllRawErrors(errs)
	return e.SetRawImage(res)
}

/*
SetRawImage takes a disgord.EmbedImage and sets the embed's image field to it, then returns the pointer to the
embed
*/
func (e *Embed) SetRawImage(img *disgord.EmbedImage) *Embed {
	e.Image = img
	return e
}

/*
SetType checks if the embed type passed to it is valid. If it is, it sets the embed's type to that, otherwise it does
nothing. It then returns the pointer to the embed
(This function fails silently)
*/
func (e *Embed) SetType(embedType string) *Embed {
	if checkTypeValid(embedType) {
		e.Type = embedType
	} else {
		e.addError(invalidEmbedTypeErrTemplateString, embedType)
	}
	return e
}
