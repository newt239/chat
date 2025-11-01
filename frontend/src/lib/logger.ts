class Logger {
  private isDevelopment = import.meta.env.DEV;

  debug(message: string, ...args: unknown[]) {
    if (this.isDevelopment) {
      console.log(`[DEBUG] ${message}`, ...args);
    }
  }

  info(message: string, ...args: unknown[]) {
    if (this.isDevelopment) {
      console.info(`[INFO] ${message}`, ...args);
    }
  }

  warn(message: string, ...args: unknown[]) {
    console.warn(`[WARN] ${message}`, ...args);
  }

  error(message: string, ...args: unknown[]) {
    console.error(`[ERROR] ${message}`, ...args);
  }
}

export const logger = new Logger();
