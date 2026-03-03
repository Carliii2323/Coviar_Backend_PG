package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"coviar_backend/internal/domain"
	"coviar_backend/internal/repository/postgres"
	"coviar_backend/pkg/config"
	"coviar_backend/pkg/database"
	"coviar_backend/pkg/validator"
)

func main() {
	fmt.Println("=== Registro de ADMINISTRADOR_APP ===")
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	// Solicitar datos de la cuenta
	fmt.Println("Datos de la cuenta:")
	email := prompt(reader, "  Email de login: ")
	password := prompt(reader, "  Password: ")

	// Validar email y password
	if err := validator.ValidateEmail(email); err != nil {
		fmt.Printf("❌ Error de email: %v\n", err)
		os.Exit(1)
	}
	if err := validator.ValidatePasswordStrength(password); err != nil {
		fmt.Printf("❌ Error de contraseña: %v\n", err)
		os.Exit(1)
	}

	// Solicitar datos del responsable
	fmt.Println("\nDatos del responsable:")
	nombre := prompt(reader, "  Nombre: ")
	apellido := prompt(reader, "  Apellido: ")
	cargo := prompt(reader, "  Cargo: ")
	dni := prompt(reader, "  DNI: ")

	if err := validator.ValidateNotEmpty(nombre, "nombre"); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
	if err := validator.ValidateNotEmpty(apellido, "apellido"); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
	if err := validator.ValidateNotEmpty(cargo, "cargo"); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
	if dni != "" {
		if err := validator.ValidateDNI(dni); err != nil {
			fmt.Printf("❌ Error de DNI: %v\n", err)
			os.Exit(1)
		}
	}

	// Inicializar DB y repositorios
	fmt.Println("\nConectando a la base de datos...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Error cargando configuración: %v\n", err)
		os.Exit(1)
	}
	db, err := database.ConnectSupabase(cfg.Supabase.URL, cfg.Supabase.Key, cfg.Supabase.DBPassword)
	if err != nil {
		fmt.Printf("❌ Error conectando a la base de datos: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	cuentaRepo := postgres.NewCuentaRepository(db.DB)
	responsableRepo := postgres.NewResponsableRepository(db.DB)

	// Generar hash de password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ Error generando hash de password: %v\n", err)
		os.Exit(1)
	}

	// Crear cuenta ADMINISTRADOR_APP
	fmt.Println("Creando cuenta...")
	cuenta := &domain.Cuenta{
		Tipo:         domain.TipoCuentaAdministradorApp,
		IDBodega:     nil,
		EmailLogin:   email,
		PasswordHash: string(hash),
	}

	idCuenta, err := cuentaRepo.Create(ctx, nil, cuenta)
	if err != nil {
		fmt.Printf("❌ Error creando cuenta: %v\n", err)
		os.Exit(1)
	}

	// Crear responsable asociado a la cuenta
	fmt.Println("Creando responsable...")
	responsable := &domain.Responsable{
		IDCuenta: idCuenta,
		Nombre:   nombre,
		Apellido: apellido,
		Cargo:    cargo,
		DNI:      dni,
		Activo:   true,
	}

	_, err = responsableRepo.Create(ctx, nil, responsable)
	if err != nil {
		fmt.Printf("❌ Error creando responsable: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Administrador creado exitosamente.\n")
	fmt.Printf("   Cuenta ID: %d\n", idCuenta)
	fmt.Printf("   Email: %s\n", email)
	fmt.Printf("   Responsable: %s %s\n", nombre, apellido)
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}
