
// #include <signal.h>

#include "cairo-dock-dock-manager.h"             // myDockObjectMgr

// local files
#include "cairo-dock-user-interaction.h"


// extern gboolean g_bUseOpenGL;
// extern gboolean g_bEasterEggs;


// TODO: To remove (1 more use in interaction):

gboolean g_bLocked;


// Those that may have to be reimplemented:

// static gint s_iNbCrashes = 0;
// static gboolean s_bSucessfulLaunch = FALSE;
// static GString *s_pLaunchCommand = NULL;
// static gint s_iLastYear = 0;
// static gboolean s_bCDSessionLaunched = FALSE; // session CD already launched?

// static gchar *s_cLastVersion = NULL;
// static gchar *s_cDefaulBackend = NULL;



// UNCHANGED
static void register_events() {
	//\___________________ register to the useful notifications.
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_DROP_DATA,
		(GldiNotificationFunc) cairo_dock_notification_drop_data,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_CLICK_ICON,
		(GldiNotificationFunc) cairo_dock_notification_click_icon,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_MIDDLE_CLICK_ICON,
		(GldiNotificationFunc) cairo_dock_notification_middle_click_icon,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_SCROLL_ICON,
		(GldiNotificationFunc) cairo_dock_notification_scroll_icon,
		GLDI_RUN_AFTER, NULL);
}




// may still need to check s_iLastYear and s_bCDSessionLaunched 
/*
// UNCHANGED
static void _cairo_dock_get_global_config (const gchar *cCairoDockDataDir)
{
	gchar *cConfFilePath = g_strdup_printf ("%s/.cairo-dock", cCairoDockDataDir);
	GKeyFile *pKeyFile = g_key_file_new ();
	if (g_file_test (cConfFilePath, G_FILE_TEST_EXISTS))
	{
		g_key_file_load_from_file (pKeyFile, cConfFilePath, 0, NULL);
		s_cLastVersion = g_key_file_get_string (pKeyFile, "Launch", "last version", NULL);
		s_cDefaulBackend = g_key_file_get_string (pKeyFile, "Launch", "default backend", NULL);
		if (s_cDefaulBackend && *s_cDefaulBackend == '\0')
		{
			g_free (s_cDefaulBackend);
			s_cDefaulBackend = NULL;
		}
		// s_iGuiMode = g_key_file_get_integer (pKeyFile, "Gui", "mode", NULL);  // 0 by default
		s_iLastYear = g_key_file_get_integer (pKeyFile, "Launch", "last year", NULL);  // 0 by default
		// s_bPingServer = g_key_file_get_boolean (pKeyFile, "Launch", "ping server", NULL);  // FALSE by default
		s_bCDSessionLaunched = g_key_file_get_boolean (pKeyFile, "Launch", "cd session", NULL);  // FALSE by default
	}
	else  // first launch or old version, the file doesn't exist yet.
	{
		gchar *cLastVersionFilePath = g_strdup_printf ("%s/.cairo-dock-last-version", cCairoDockDataDir);
		if (g_file_test (cLastVersionFilePath, G_FILE_TEST_EXISTS))
		{
			gsize length = 0;
			g_file_get_contents (cLastVersionFilePath,
				&s_cLastVersion,
				&length,
				NULL);
		}
		g_remove (cLastVersionFilePath);
		g_free (cLastVersionFilePath);
		g_key_file_set_string (pKeyFile, "Launch", "last version", s_cLastVersion?s_cLastVersion:"");
		
		g_key_file_set_string (pKeyFile, "Launch", "default backend", "");
		
		// g_key_file_set_integer (pKeyFile, "Gui", "mode", s_iGuiMode);
		
		g_key_file_set_integer (pKeyFile, "Launch", "last year", s_iLastYear);
		
		cairo_dock_write_keys_to_file (pKeyFile, cConfFilePath);
	}
	g_key_file_free (pKeyFile);
	g_free (cConfFilePath);
}
*/
