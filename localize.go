package main

import "fmt"

// Simple localization approach for top-level log messages
func locMsg(key, lang string) string {
	messages := map[string]map[string]string{
		"start_organizer": {
			"en": "=== Started File Organizer at %s ===",
			"es": "=== Iniciando el organizador de archivos en %s ===",
		},
		"input_folder": {
			"en": "Input folder: %s",
			"es": "Carpeta de entrada: %s",
		},
		"output_folder": {
			"en": "Output folder: %s",
			"es": "Carpeta de salida: %s",
		},
		"input_folder_invalid": {
			"en": "Input folder check failed",
			"es": "Error al verificar la carpeta de entrada",
		},
		"error_organizing": {
			"en": "Error organizing files",
			"es": "Error organizando archivos",
		},
		"file_org_complete": {
			"en": "File organization complete.",
			"es": "Organización de archivos completa.",
		},
		"finished": {
			"en": "=== Finished at %s ===",
			"es": "=== Finalizado a las %s ===",
		},
		"skipping_file": {
			"en": "Skipping file already in output folder: %s",
			"es": "Saltando archivo, ya se encuentra en carpeta de salida: %s",
		},
		"move_error": {
			"en": "Error moving file %q to %q: %v",
			"es": "Error al mover archivo %q a %q: %v",
		},
		"moved_file": {
			"en": "Moved: %q => %q",
			"es": "Movido: %q => %q",
		},
	}

	// Fallback logic: if the key or lang is missing, default to English
	if msgMap, ok := messages[key]; ok {
		if msg, ok := msgMap[lang]; ok {
			return msg
		}
		return msgMap["en"]
	}
	// If the key is unknown, fallback to a simple message in English
	return fmt.Sprintf("Missing translation for key=%q", key)
}
