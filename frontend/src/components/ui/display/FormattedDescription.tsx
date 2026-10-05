interface FormattedDescriptionProps {
  text: string;
}

export const FormattedDescription = ({ text }: FormattedDescriptionProps) => {
  const parts = text.split(/\*\*(.*?)\*\*/);
  return (
    <span className="whitespace-pre-line">
      {parts.map((part, i) => (i % 2 === 1 ? <strong key={i}>{part}</strong> : part))}
    </span>
  );
};
