/*
Package wizard provides support for field based forms defined by client code. A key-value storage is used to save
the state of a form between messages.

To add a form to your [github.com/kozalosev/SadFavBot/base.MessageHandler], it must implement the [WizardMessageHandler]
interface and create a [Wizard] in its [github.com/kozalosev/SadFavBot/base.MessageHandler.Handle] method.
*/
package wizard
