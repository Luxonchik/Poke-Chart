import { cn } from "@/lib/utils";

interface ContainerProps {
  children: React.ReactNode;
  className?: string;
}

/** Page-wide horizontal bounds shared by the header, full-width bands and page content. */
export function Container({ children, className }: ContainerProps) {
  return <div className={cn("mx-auto w-full max-w-[1600px] px-6 lg:px-12", className)}>{children}</div>;
}
