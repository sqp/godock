
#include <glib/gstdio.h>  // g_mkdir/g_remove

#include "cairo-dock-applications-manager.h"       // myTaskbarParam
#include "cairo-dock-class-manager.h"              // cairo_dock_get_class_icon
#include "cairo-dock-dialog-factory.h"             // gldi_dialog_show_temporary_with_default_icon
#include "cairo-dock-dock-manager.h"               // myDocksParam
#include "cairo-dock-file-manager.h"               // cairo_dock_copy_file
#include "cairo-dock-icon-facility.h"              // cairo_dock_get_icon_order
#include "cairo-dock-launcher-manager.h"           // gldi_launcher_add_new
#include "cairo-dock-stack-icon-manager.h"         // CAIRO_DOCK_ICON_TYPE_IS_CONTAINER
#include "cairo-dock-themes-manager.h"             // cairo_dock_update_conf_file
#include "cairo-dock-utils.h"                      // cairo_dock_launch_command_sync


#include "cairo-dock-log.h"



extern CairoDock *g_pMainDock;

extern gchar* g_cCurrentIconsPath;
extern gchar* g_cConfFile;


// Craps to remake.

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



