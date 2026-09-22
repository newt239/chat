class Logger {
  private readonly isDevelopment = import.meta.env.DEV;

  public debug(message: string, ...args: unknown[]) {
    if (this.isDevelopment) {
      console.log(`[DEBUG] ${message}`, ...args);
    }
  }

  public info(message: string, ...args: unknown[]) {
    if (this.isDevelopment) {
      console.info(`[INFO] ${message}`, ...args);
    }
  }

  public warn(message: string, ...args: unknown[]) {
    console.warn(`[WARN] ${message}`, ...args);
  }

  public error(message: string, ...args: unknown[]) {
    console.error(`[ERROR] ${message}`, ...args);
  }
}

export const logger = new Logger();
