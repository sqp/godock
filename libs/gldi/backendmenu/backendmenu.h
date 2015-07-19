
#include <math.h>

#include <glib/gstdio.h>  // g_mkdir/g_remove

#include "gldi-icon-names.h"                       // GLDI_ICON_NAME_*

#include "cairo-dock-applications-manager.h"       // myTaskbarParam
#include "cairo-dock-class-manager.h"              // cairo_dock_get_class_icon
#include "cairo-dock-desktop-manager.h"            // gldi_desktop_get_width
#include "cairo-dock-dialog-factory.h"       // gldi_dialog_show_temporary_with_default_icon
#include "cairo-dock-dock-manager.h"       // 
#include "cairo-dock-file-manager.h"              // cairo_dock_copy_file
#include "cairo-dock-icon-facility.h"              // cairo_dock_get_icon_order
#include "cairo-dock-launcher-manager.h"           // GLDI_OBJECT_IS_LAUNCHER_ICON
#include "cairo-dock-stack-icon-manager.h"         // GLDI_OBJECT_IS_STACK_ICON
#include "cairo-dock-themes-manager.h"         // cairo_dock_update_conf_file
#include "cairo-dock-utils.h"                      // cairo_dock_launch_command_sync


#include "cairo-dock-log.h"



extern CairoDock *g_pMainDock;

extern gchar* g_cCurrentIconsPath;
extern gchar* g_cConfFile;


// Craps to remake.

#define CAIRO_DOCK_ABOUT_WIDTH 400
#define CAIRO_DOCK_ABOUT_HEIGHT 500
#define CAIRO_DOCK_FILE_HOST_URL "https://launchpad.net/cairo-dock"  // https://developer.berlios.de/project/showfiles.php?group_id=8724
#define CAIRO_DOCK_SITE_URL "http://glx-dock.org"  // http://cairo-dock.vef.fr
#define CAIRO_DOCK_FORUM_URL "http://forum.glx-dock.org"  // http://cairo-dock.vef.fr/bg_forumlist.php
#define CAIRO_DOCK_PAYPAL_URL "https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=UWQ3VVRB2ZTZS&lc=GB&item_name=Support%20Cairo%2dDock&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donate_LG%2egif%3aNonHosted"
#define CAIRO_DOCK_FLATTR_URL "http://flattr.com/thing/370779/Support-Cairo-Dock-development"

#define CAIRO_DOCK_LOGO "cairo-dock-logo.png"

#define CAIRO_DOCK_PLUGINS_EXTRAS_URL "http://extras.glx-dock.org"
// #define DISTANT_DIR "3.4.0"

// static gchar *cairo_dock_get_third_party_applets_link (void)
// {
// 	return g_strdup_printf (CAIRO_DOCK_PLUGINS_EXTRAS_URL"/"DISTANT_DIR);
// }


static void _cairo_dock_add_about_page (GtkWidget *pNoteBook, const gchar *cPageLabel, const gchar *cAboutText)
{
	GtkWidget *pVBox, *pScrolledWindow;
	GtkWidget *pPageLabel, *pAboutLabel;
	
	pPageLabel = gtk_label_new (cPageLabel);
	pVBox = gtk_box_new (GTK_ORIENTATION_VERTICAL, 0);
	pScrolledWindow = gtk_scrolled_window_new (NULL, NULL);
	gtk_scrolled_window_set_policy (GTK_SCROLLED_WINDOW (pScrolledWindow), GTK_POLICY_AUTOMATIC, GTK_POLICY_AUTOMATIC);
	#if GTK_CHECK_VERSION (3, 8, 0)
	gtk_container_add (GTK_CONTAINER (pScrolledWindow), pVBox);
	#else
	gtk_scrolled_window_add_with_viewport (GTK_SCROLLED_WINDOW (pScrolledWindow), pVBox);
	#endif
	gtk_notebook_append_page (GTK_NOTEBOOK (pNoteBook), pScrolledWindow, pPageLabel);
	
	pAboutLabel = gtk_label_new (NULL);
	gtk_label_set_use_markup (GTK_LABEL (pAboutLabel), TRUE);

// GTK WARNINGS
	// gtk_misc_set_alignment (GTK_MISC (pAboutLabel), 0.0, 0.0);
	// gtk_misc_set_padding (GTK_MISC (pAboutLabel), 30, 0);
	

	gtk_box_pack_start (GTK_BOX (pVBox),
		pAboutLabel,
		FALSE,
		FALSE,
		15);
	gtk_label_set_markup (GTK_LABEL (pAboutLabel), cAboutText);
}

static void _cairo_dock_about (GldiContainer *pContainer)
{
	// build dialog
	GtkWidget *pDialog = gtk_dialog_new_with_buttons (_("About Cairo-Dock"),
		GTK_WINDOW (pContainer->pWidget),
		GTK_DIALOG_DESTROY_WITH_PARENT,
		GLDI_ICON_NAME_CLOSE,
		GTK_RESPONSE_CLOSE,
		NULL);

	// the dialog box is destroyed when the user responds
	g_signal_connect_swapped (pDialog,
		"response",
		G_CALLBACK (gtk_widget_destroy),
		pDialog);

	GtkWidget *pContentBox = gtk_dialog_get_content_area (GTK_DIALOG(pDialog));

	// logo + links
	GtkWidget *pHBox = gtk_box_new (GTK_ORIENTATION_HORIZONTAL, 0);
	gtk_box_pack_start (GTK_BOX (pContentBox), pHBox, FALSE, FALSE, 0);

	const gchar *cImagePath = GLDI_SHARE_DATA_DIR"/images/"CAIRO_DOCK_LOGO;
	GtkWidget *pImage = gtk_image_new_from_file (cImagePath);
	gtk_box_pack_start (GTK_BOX (pHBox), pImage, FALSE, FALSE, 0);

	GtkWidget *pVBox = gtk_box_new (GTK_ORIENTATION_VERTICAL, 0);
	gtk_box_pack_start (GTK_BOX (pHBox), pVBox, FALSE, FALSE, 0);
	
	GtkWidget *pLink = gtk_link_button_new_with_label (CAIRO_DOCK_SITE_URL, "Cairo-Dock (2007-2014)\n version "GLDI_VERSION);
	gtk_box_pack_start (GTK_BOX (pVBox), pLink, FALSE, FALSE, 0);
	
	//~ pLink = gtk_link_button_new_with_label (CAIRO_DOCK_FORUM_URL, _("Community site"));
	//~ gtk_widget_set_tooltip_text (pLink, _("Problems? Suggestions? Just want to talk to us? Come on over!"));
	//~ gtk_box_pack_start (GTK_BOX (pVBox), pLink, FALSE, FALSE, 0);
	
	pLink = gtk_link_button_new_with_label (CAIRO_DOCK_FILE_HOST_URL, _("Development site"));
	gtk_widget_set_tooltip_text (pLink, _("Find the latest version of Cairo-Dock here !"));
	gtk_box_pack_start (GTK_BOX (pVBox), pLink, FALSE, FALSE, 0);
	
	// gchar *cLink = cairo_dock_get_third_party_applets_link ();
	// pLink = gtk_link_button_new_with_label (cLink, _("Get more applets!"));
	// g_free (cLink);
	// gtk_box_pack_start (GTK_BOX (pVBox), pLink, FALSE, FALSE, 0);
	
	gchar *cLabel = g_strdup_printf ("%s (Flattr)", _("Donate"));
	pLink = gtk_link_button_new_with_label (CAIRO_DOCK_FLATTR_URL, cLabel);
	g_free (cLabel);
	gtk_widget_set_tooltip_text (pLink, _("Support the people who spend countless hours to bring you the best dock ever."));
	gtk_box_pack_start (GTK_BOX (pVBox), pLink, FALSE, FALSE, 0);
	
	cLabel = g_strdup_printf ("%s (Paypal)", _("Donate"));
	pLink = gtk_link_button_new_with_label (CAIRO_DOCK_PAYPAL_URL, cLabel);
	g_free (cLabel);
	gtk_widget_set_tooltip_text (pLink, _("Support the people who spend countless hours to bring you the best dock ever."));
	gtk_box_pack_start (GTK_BOX (pVBox), pLink, FALSE, FALSE, 0);
	
	
	// notebook
	GtkWidget *pNoteBook = gtk_notebook_new ();
	gtk_notebook_set_scrollable (GTK_NOTEBOOK (pNoteBook), TRUE);
	gtk_notebook_popup_enable (GTK_NOTEBOOK (pNoteBook));
	gtk_box_pack_start (GTK_BOX (pContentBox), pNoteBook, TRUE, TRUE, 0);

	// About
	/* gchar *text = g_strdup_printf ("\n\n<b>%s</b>\n\n\n"
		"<a href=\"http://glx-dock.org\">http://glx-dock.org</a>",
		_("<b>Cairo-Dock is a pretty, light and convenient interface\n"
			" to your desktop, able to replace advantageously your system panel!</b>"));
	_cairo_dock_add_about_page (pNoteBook,
		_("About"),
		text);*/
	// Development
	gchar *text = g_strdup_printf ("%s\n\n"
	"<span size=\"larger\" weight=\"bold\">%s</span>\n\n"
		"  Fabounet (Fabrice Rey)\n"
		"\t<span size=\"smaller\">%s</span>\n\n"
		"  Matttbe (Matthieu Baerts)\n"
		"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n"
		"  Eduardo Mucelli\n"
		"  Jesuisbenjamin\n"
		"  SQP\n",
		_("Here is a list of the current developers and contributors"),
		_("Developers"),
		_("Main developer and project leader"),
		_("Contributors / Hackers"));
	_cairo_dock_add_about_page (pNoteBook,
		_("Development"),
		text);
	// Support
		text = g_strdup_printf ("<span size=\"larger\" weight=\"bold\">%s</span>\n\n"
		"  Matttbe\n"
		"  Mav\n"
		"  Necropotame\n"
		"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n"
		"  BobH\n"
		"  Franksuse64\n"
		"  Lylambda\n"
		"  Ppmt\n"
		"  Taiebot65\n"
		"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n"
		"%s",
		_("Website"),
		_("Beta-testing / Suggestions / Forum animation"),
		_("Translators for this language"),
		_("translator-credits"));
	_cairo_dock_add_about_page (pNoteBook,
		_("Support"),
		text);
	// Thanks
		text = g_strdup_printf ("%s\n"
		"<a href=\"http://glx-dock.org/ww_page.php?p=How to help us\">%s</a>: %s\n\n"
		"\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n"
		"  Augur\n"
		"  ChAnGFu\n"
		"  Ctaf\n"
		"  Mav\n"
		"  Necropotame\n"
		"  Nochka85\n"
		"  Paradoxxx_Zero\n"
		"  Rom1\n"
		"  Tofe\n"
		"  Mac Slow (original idea)\n"
		"\t<span size=\"smaller\">%s</span>\n"
		"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n"
		"\t<a href=\"http://glx-dock.org/userlist_messages.php\">%s</a>\n"
		"\n\n<span size=\"larger\" weight=\"bold\">%s</span>\n\n"
		"  Benoit2600\n"
		"  Coz\n"
		"  Fabounet\n"
		"  Lord Northam\n"
		"  Lylambda\n"
		"  MastroPino\n"
		"  Matttbe\n"
		"  Nochka85\n"
		"  Paradoxxx_Zero\n"
		"  Taiebot65\n",
		_("Thanks to all people that help us to improve the Cairo-Dock project.\n"
			"Thanks to all current, former and future contributors."),
		_("How to help us?"),
		_("Don't hesitate to join the project, we need you ;)"),
		_("Former contributors"),
		_("For a complete list, please have a look to BZR logs"),
		_("Users of our forum"),
		_("List of our forum's members"),
		_("Artwork"));
	_cairo_dock_add_about_page (pNoteBook,
		_("Thanks"),
		text);
	g_free (text);
	
	gtk_window_resize (GTK_WINDOW (pDialog),
		MIN (CAIRO_DOCK_ABOUT_WIDTH, gldi_desktop_get_width()),
		MIN (CAIRO_DOCK_ABOUT_HEIGHT, gldi_desktop_get_height() - (g_pMainDock && g_pMainDock->container.bIsHorizontal ? g_pMainDock->iMaxDockHeight : 0)));

	gtk_widget_show_all (pDialog);

	gtk_window_set_keep_above (GTK_WINDOW (pDialog), TRUE);
	//don't use gtk_dialog_run(), as we don't want to block the dock
}


//
//------------------------------------------------------------[ CUSTOM ICONS ]--


static void _cairo_dock_make_launcher_from_appli (Icon *icon, CairoDock *pDock)
{
	
	g_return_if_fail (icon->cClass != NULL);
	
	// look for the .desktop file of the program
	cd_debug ("%s (%s)", __func__, icon->cClass);
	gchar *cDesktopFilePath = g_strdup (cairo_dock_get_class_desktop_file (icon->cClass));
	if (cDesktopFilePath == NULL)  // empty class
	{
		gchar *cCommand = g_strdup_printf ("find /usr/share/applications /usr/local/share/applications -iname \"*%s*.desktop\"", icon->cClass);  // look for a desktop file from their file name
		gchar *cResult = cairo_dock_launch_command_sync (cCommand);
		if (cResult == NULL || *cResult == '\0')  // no luck, search harder
		{
			g_free (cCommand);
			cCommand = g_strdup_printf ("find /usr/share/applications /usr/local/share/applications -name \"*.desktop\" -exec grep -qi '%s' {} \\; -print", icon->cClass);  // look for a desktop file from their content
			cResult = cairo_dock_launch_command_sync (cCommand);
		}
		if (cResult != NULL && *cResult != '\0')
		{
			gchar *str = strchr (cResult, '\n');  // remove the trailing linefeed, and only take the first result
			if (str)
				*str = '\0';
			cDesktopFilePath = cResult;
		}
		g_free (cCommand);
	}
	
	// make a new launcher from this desktop file
	if (cDesktopFilePath != NULL)
	{
		cd_message ("found desktop file : %s", cDesktopFilePath);
		// place it after the last launcher, since the user will probably want to move this new launcher amongst the already existing ones.
		double fOrder = CAIRO_DOCK_LAST_ORDER;
		Icon *pIcon;
		GList *ic, *last_launcher_ic = NULL;
		for (ic = g_pMainDock->icons; ic != NULL; ic = ic->next)
		{
			pIcon = ic->data;
			if (CAIRO_DOCK_ICON_TYPE_IS_LAUNCHER (pIcon)
			|| CAIRO_DOCK_ICON_TYPE_IS_CONTAINER (pIcon))
			{
				last_launcher_ic = ic;
			}
		}
		if (last_launcher_ic != NULL)
		{
			ic = last_launcher_ic;
			pIcon = ic->data;
			Icon *next_icon = (ic->next ? ic->next->data : NULL);
			if (next_icon != NULL && cairo_dock_get_icon_order (next_icon) == cairo_dock_get_icon_order (pIcon))
				fOrder = (pIcon->fOrder + next_icon->fOrder) / 2;
			else
				fOrder = pIcon->fOrder + 1;
		}
		gldi_launcher_add_new (cDesktopFilePath, g_pMainDock, fOrder);  // add in the main dock
	}
	else
	{
		gldi_dialog_show_temporary_with_default_icon (_("Sorry, couldn't find the corresponding description file.\nConsider dragging and dropping the launcher from the Applications Menu."), icon, CAIRO_CONTAINER (pDock), 8000);
	}
	g_free (cDesktopFilePath);
	
}


//
//------------------------------------------------------------[ CUSTOM ICONS ]--

static void cairo_dock_set_custom_icon_on_appli (const gchar *cFilePath, Icon *icon, GldiContainer *pContainer)
{
	g_return_if_fail (CAIRO_DOCK_IS_APPLI (icon) && cFilePath != NULL);
	gchar *ext = strrchr (cFilePath, '.');
	if (!ext)
		return;
	cd_debug ("%s (%s - %s)", __func__, cFilePath, icon->cFileName);
	if ((strcmp (ext, ".png") == 0 || strcmp (ext, ".svg") == 0) && !myDocksParam.bLockAll) // && ! myDocksParam.bLockIcons) // or if we have to hide the option...
	{
		if (!myTaskbarParam.bOverWriteXIcons)
		{
			myTaskbarParam.bOverWriteXIcons = TRUE;
			cairo_dock_update_conf_file (g_cConfFile,
				G_TYPE_BOOLEAN, "TaskBar", "overwrite xicon", myTaskbarParam.bOverWriteXIcons,
				G_TYPE_INVALID);
			gldi_dialog_show_temporary_with_default_icon (_("The option 'overwrite X icons' has been automatically enabled in the config.\nIt is located in the 'Taskbar' module."), icon, pContainer, 6000);
		}
		
		gchar *cPath = NULL;
		if (strncmp (cFilePath, "file://", 7) == 0)
		{
			cPath = g_filename_from_uri (cFilePath, NULL, NULL);
		}
		
		const gchar *cClassIcon = cairo_dock_get_class_icon (icon->cClass);
		if (cClassIcon == NULL)
			cClassIcon = icon->cClass;
		
		gchar *cDestPath = g_strdup_printf ("%s/%s%s", g_cCurrentIconsPath, cClassIcon, ext);
		cairo_dock_copy_file (cPath?cPath:cFilePath, cDestPath);
		g_free (cDestPath);
		g_free (cPath);
		
		cairo_dock_reload_icon_image (icon, pContainer);
		cairo_dock_redraw_icon (icon);
	}
}



static void _show_image_preview (GtkFileChooser *pFileChooser, GtkImage *pPreviewImage)
{
	gchar *cFileName = gtk_file_chooser_get_preview_filename (pFileChooser);
	if (cFileName == NULL)
		return ;
	GdkPixbuf *pixbuf = gdk_pixbuf_new_from_file_at_size (cFileName, 64, 64, NULL);
	g_free (cFileName);
	if (pixbuf != NULL)
	{
		gtk_image_set_from_pixbuf (pPreviewImage, pixbuf);
		g_object_unref (pixbuf);
		gtk_file_chooser_set_preview_widget_active (pFileChooser, TRUE);
	}
	else
		gtk_file_chooser_set_preview_widget_active (pFileChooser, FALSE);
}

static void _cairo_dock_set_custom_appli_icon (Icon *icon, CairoDock *pDock)
{
	if (! CAIRO_DOCK_IS_APPLI (icon))
		return;
	
	GtkWidget* pFileChooserDialog = gtk_file_chooser_dialog_new (
		_("Pick up an image"),
		GTK_WINDOW (pDock->container.pWidget),
		GTK_FILE_CHOOSER_ACTION_OPEN,
		_("Ok"),
		GTK_RESPONSE_OK,
		_("Cancel"),
		GTK_RESPONSE_CANCEL,
		NULL);
	gtk_file_chooser_set_current_folder (GTK_FILE_CHOOSER (pFileChooserDialog), "/usr/share/icons");  // we could also use 'xdg-user-dir PICTURES' or /usr/share/icons or ~/.icons, but actually we have no idea where the user will want to pick the image, so let's not try to be smart.
	gtk_file_chooser_set_select_multiple (GTK_FILE_CHOOSER (pFileChooserDialog), FALSE);
	
	GtkWidget *pPreviewImage = gtk_image_new ();
	gtk_file_chooser_set_preview_widget (GTK_FILE_CHOOSER (pFileChooserDialog), pPreviewImage);
	g_signal_connect (GTK_FILE_CHOOSER (pFileChooserDialog), "update-preview", G_CALLBACK (_show_image_preview), pPreviewImage);

	// a filter
	GtkFileFilter *pFilter = gtk_file_filter_new ();
	gtk_file_filter_set_name (pFilter, _("Image"));
	gtk_file_filter_add_pixbuf_formats (pFilter);
	gtk_file_chooser_add_filter (GTK_FILE_CHOOSER (pFileChooserDialog), pFilter);
	
	gtk_widget_show (pFileChooserDialog);
	int answer = gtk_dialog_run (GTK_DIALOG (pFileChooserDialog));
	if (answer == GTK_RESPONSE_OK)
	{
		if (myTaskbarParam.cOverwriteException != NULL && icon->cClass != NULL)  // si cette classe etait definie pour ne pas ecraser l'icone X, on le change, sinon l'action utilisateur n'aura aucun impact, ce sera troublant.
		{
			gchar **pExceptions = g_strsplit (myTaskbarParam.cOverwriteException, ";", -1);
			
			int i, j = -1;
			for (i = 0; pExceptions[i] != NULL; i ++)
			{
				if (j == -1 && strcmp (pExceptions[i], icon->cClass) == 0)
				{
					g_free (pExceptions[i]);
					pExceptions[i] = NULL;
					j = i;
				}
			}  // apres la boucle, i = nbre d'elements, j = l'element qui a ete enleve.
			if (j != -1)  // un element a ete enleve.
			{
				cd_warning ("The class '%s' was explicitely set up to use the X icon, we'll change this behavior automatically.", icon->cClass);
				if (j < i - 1)  // ce n'est pas le dernier
				{
					pExceptions[j] = pExceptions[i-1];
					pExceptions[i-1] = NULL;
				}
				
				myTaskbarParam.cOverwriteException = g_strjoinv (";", pExceptions);
				cairo_dock_set_overwrite_exceptions (myTaskbarParam.cOverwriteException);
				
				cairo_dock_update_conf_file (g_cConfFile,
					G_TYPE_STRING, "TaskBar", "overwrite exception", myTaskbarParam.cOverwriteException,
					G_TYPE_INVALID);
			}
			g_strfreev (pExceptions);
		}
		
		gchar *cFilePath = gtk_file_chooser_get_filename (GTK_FILE_CHOOSER (pFileChooserDialog));
		cairo_dock_set_custom_icon_on_appli (cFilePath, icon, CAIRO_CONTAINER (pDock));
		g_free (cFilePath);
	}
	gtk_widget_destroy (pFileChooserDialog);
}

static void _cairo_dock_remove_custom_appli_icon (Icon *icon, CairoDock *pDock)
{
	if (! CAIRO_DOCK_IS_APPLI (icon))
		return;
	
	const gchar *cClassIcon = cairo_dock_get_class_icon (icon->cClass);
	if (cClassIcon == NULL)
		cClassIcon = icon->cClass;
	
	gchar *cCustomIcon = g_strdup_printf ("%s/%s.png", g_cCurrentIconsPath, cClassIcon);
	if (!g_file_test (cCustomIcon, G_FILE_TEST_EXISTS))
	{
		g_free (cCustomIcon);
		cCustomIcon = g_strdup_printf ("%s/%s.svg", g_cCurrentIconsPath, cClassIcon);
		if (!g_file_test (cCustomIcon, G_FILE_TEST_EXISTS))
		{
			g_free (cCustomIcon);
			cCustomIcon = NULL;
		}
	}
	if (cCustomIcon != NULL)
	{
		g_remove (cCustomIcon);
		cairo_dock_reload_icon_image (icon, CAIRO_CONTAINER (pDock));
		cairo_dock_redraw_icon (icon);
	}
}



